package main

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/jszwec/csvutil"
	"github.com/pkg/errors"
)

func convert(fileName string, delimiter rune, tagFormats []FormatTag, mode SortingMode) error {
	splitedName := strings.Split(fileName, ".")
	if len(splitedName) != 2 {
		return errors.Errorf("incorrect file name: %v", fileName)
	}

	filePrefix := splitedName[0] + "."

	hotels, err := readFile(fileName, delimiter)
	if err != nil {
		return errors.Wrapf(err, "read file %v", fileName)
	}

	if len(mode) != 0 {
		hotels = sortHotels(hotels, mode)
	}

	for _, tag := range tagFormats {
		bytes, err := createMarshalBody(hotels, tag)
		if err != nil {
			log.Print(errors.Wrapf(err, "marshal body for tag: %v", tag))
			continue
		}

		filePath := filePrefix + tag.String()
		if err := ioutil.WriteFile(filePath, bytes, 0655); err != nil {
			log.Print(errors.Wrapf(err, "save file by filePath: %v", filePath))
			continue
		}
	}

	return nil
}

func readFile(fileName string, delimiter rune) ([]ConvertedHotelInfo, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, errors.Wrap(err, "opening file")
	}

	if !strings.Contains(fileName, csvTag.String()) {
		return nil, errors.Errorf("unsupported file for reading: %v", fileName)
	}

	hotels, warns, err := readCSV(file, delimiter)
	if err != nil {
		return nil, errors.Wrapf(err, "read info from file")
	} else if warns > 0 {
		log.Printf("There are %v warns during reading file %v", warns, fileName)
	}

	log.Printf("%v hotels was read", len(hotels))
	return hotels, nil
}

func readCSV(r io.Reader, delimiter rune) ([]ConvertedHotelInfo, int, error) {
	csvReader := csv.NewReader(r)
	csvReader.Comma = delimiter

	dec, err := csvutil.NewDecoder(csvReader)
	if err != nil {
		return nil, 0, errors.Wrap(err, "creation csv util decoder")
	}

	var (
		hotels = make([]ConvertedHotelInfo, 0)
		warns  int
	)

	for {
		hotel := new(HotelsInfo)
		if err := dec.Decode(hotel); err == io.EOF {
			break
		} else if err != nil {
			return nil, 0, errors.Wrap(err, "decode struct")
		}

		if !isValid(*hotel) {
			warns++
			continue
		}

		hotels = append(hotels, getConvertedHotelInfo(*hotel))
	}

	return hotels, warns, nil
}

func isValid(hotel HotelsInfo) bool {
	u, err := url.Parse(hotel.URI)
	if err != nil {
		log.Println(errors.Wrapf(err, "parse hotel URI: %v", hotel.URI))
		return false
	}

	switch {
	case len(u.Scheme) == 0 || len(u.Host) == 0:
		return false
	case hotel.Stars < 0 || hotel.Stars > 5:
		return false
	case !utf8.Valid([]byte(hotel.HotelName)):
		return false
	}

	return true
}

func getConvertedHotelInfo(hotel HotelsInfo) ConvertedHotelInfo {
	return ConvertedHotelInfo{
		HotelName: hotel.HotelName,
		Address:   hotel.Address,
		Stars:     hotel.Stars,
		Contact:   hotel.Contact,
		Phone:     hotel.Phone,
		URI:       hotel.URI,
	}
}

func sortHotels(hotels []ConvertedHotelInfo, sortingMode SortingMode) []ConvertedHotelInfo {
	switch sortingMode {
	case sortByNames:
		sort.Slice(hotels, func(i, j int) bool {
			return strings.ToLower(hotels[i].HotelName) < strings.ToLower(hotels[j].HotelName)
		})
	case sortByStars:
		sort.Slice(hotels, func(i, j int) bool { return hotels[i].Stars > hotels[j].Stars })
	}

	return hotels
}

func createMarshalBody(hotels []ConvertedHotelInfo, tag FormatTag) ([]byte, error) {
	var (
		body []byte
		err  error
	)

	switch tag {
	case xmlTag:
		body, err = xml.Marshal(&hotels)
	case jsonTag:
		body, err = json.Marshal(&hotels)
	default:
		return nil, errors.Errorf("unsupported tag: %v", tag)
	}

	if err != nil {
		return nil, errors.Wrap(err, "marshal struct")
	}

	return body, nil
}
