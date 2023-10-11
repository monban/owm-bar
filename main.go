package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	owm "github.com/briandowns/openweathermap"
)

func getWeatherIcon(id int, daytime bool) rune {
	// Icons that don't change based on time of day
	switch id {
	case 711: // Smoke
		return '\ue35c'
	case 721, 741: // Haze, fog
		return '\ue313'
	case 761: // Dust
		return '\ue35d'
	case 762: // Volcanic dust
		return '\ue3c0'
	case 781: // Tornado
		return '\ue351'
	case 804: // Complete cloud cover
		return '\ue312'
	}

	if daytime { // Day
		switch {
		case id >= 200 && id < 300: // Thunderstorm
			return '\ue30f'
		case id >= 300 && id < 400: // Drizzle
			return '\ue30b'
		case id >= 500 && id < 600: // Rain
			return '\ue308'
		case id >= 600 && id < 700: // Snow
			return '\ue30a'
		case id == 800: // Clear
			return '\ue30d'
		case id == 801: // Few clouds
			return '\ue30c'
		case id == 802 || id == 803: // Cloudy
			return '\ue302'
		}
	} else { // Night
		switch {
		case id >= 200 && id < 300: // Thunderstorm
			return '\ue338'
		case id >= 300 && id < 400: // Drizzle
			return '\ue334'
		case id >= 500 && id < 600: // Rain
			return '\ue333'
		case id >= 600 && id < 700: // Snow
			return '\ue335'
		case id == 800: // Clear
			return '\ue32b'
		case id == 801: // Few clouds
			return '\ue37b'
		case id == 802 || id == 803: // Cloudy
			return '\ue32e'
		}
	}

	// No matches
	return '\uFFFD'
}

type owmKey string

func (k *owmKey) Set(v string) error {
	if len(v) < 1 {
		return fmt.Errorf("key must be at least one character")
	}
	*k = owmKey(v)
	return nil
}

func (k owmKey) String() string {
	return string(k)
}

type Coordinates owm.Coordinates

func (coords *Coordinates) Set(v string) error {
	_, err := fmt.Sscanf(v, "%f,%f", &coords.Latitude, &coords.Longitude)
	if err != nil {
		return fmt.Errorf("Error parsing coordinates: %s", err.Error())
	}

	return nil
}

func (coords *Coordinates) String() string {
	return fmt.Sprintf("Latitude: %f, Longitude: %f", coords.Latitude, coords.Longitude)
}

func isDaytime(sys owm.Sys) bool {
	t := int(time.Now().Unix())
	if t > sys.Sunrise && t < sys.Sunset {
		return true
	}
	return false
}

func main() {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	key := owmKey("")
	coords := Coordinates{}

	fs.Var(&key, "key", "Open Weather Map API key")
	fs.Var(&coords, "coords", "GCS coordinates (example: 51,0)")
	units := fs.String("units", "C", "temperature units (K, C, or F)")
	lang := fs.String("lang", "en", "language in 639-1 code (examples: en, la, fr)")
	debug := fs.Bool("debug", false, "will print out the complete text returned from the server")
	useIcon := fs.Bool("icon", false, "use NerdFont icon instead of text description")

	fs.Parse(os.Args[1:])

	if key == "" {
		fmt.Println("must specify API key, see --help for usage")
		return
	}

	if coords.Latitude == 0 && coords.Longitude == 0 {
		fmt.Println("must specify coordinates, see --help for usage")
		return
	}

	result, err := owm.NewCurrent(*units, *lang, string(key))
	if err != nil {
		fmt.Println(err)
		return
	}
	owmCoords := owm.Coordinates(coords)
	err = result.CurrentByCoordinates(&owmCoords)
	if err != nil {
		fmt.Println(err)
		return
	}
	if *debug {
		fmt.Printf("%+v\n", result)
	}

	weatherWords := make([]string, len(result.Weather))
	for i, v := range result.Weather {
		weatherWords[i] = v.Description
	}
	description := strings.Join(weatherWords, " ")
	_ = description
	var icon rune
	if *useIcon {
		icon = getWeatherIcon(result.Weather[0].ID, isDaytime(result.Sys))
	}

	fmt.Printf("%c %.1f°%s", icon, result.Main.FeelsLike, *units)
}
