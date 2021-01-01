package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"github.com/stretchr/testify/assert"
)

type CaseConvert struct {
	Number       int
	CsvFile      string
	ExpectedFile []byte
}

func TestFunc(t *testing.T) {
	assrt := assert.New(t)
	for _, c := range getMainCases() {
		_ = convert(c.CsvFile, ',', []FormatTag{jsonTag}, "")
		body, _ := openAndReadFile("hotels_test." + jsonTag.String())

		assrt.Equal(c.ExpectedFile, body, "Test number: %v", c.Number)
	}
}

var expectedJSOn = []byte(`[{"name":"The Gibson","address":"63847 Lowe Knoll, East Maxine, WA 97030-4876","stars":5,"contact":"Dr. Sinda Wyman","phone":"1-270-665-9933x1626","uri":"http://www.paucek.com/search.htm"},{"name":"Martini Cattaneo","address":"Stretto Bernardi 004, Quarto Mietta nell'emilia, 07958 Torino (OG)","stars":5,"contact":"Rosalino Marchetti","phone":"+39 627 68225719","uri":"http://www.farina.org/blog/categories/tags/about.html"}]`)

func openAndReadFile(fileName string) ([]byte, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(file)
}

func getMainCases() []CaseConvert {
	return []CaseConvert{
		{
			Number:       0,
			CsvFile:      "hotels_test.csv",
			ExpectedFile: expectedJSOn,
		},
	}
}

type CaseValidation struct {
	Number     int
	HotelsInfo HotelsInfo
	Expected   bool
}

func TestValidation(t *testing.T) {
	assrt := assert.New(t)
	for _, c := range getValidationCases() {
		assrt.Equal(c.Expected, isValid(c.HotelsInfo), "Test number: %v", c.Number)
	}
}

func getValidationCases() []CaseValidation {
	s := "今日は"
	var b bytes.Buffer
	wInUTF8 := transform.NewWriter(&b, japanese.ShiftJIS.NewEncoder())
	_, _ = wInUTF8.Write([]byte(s))
	_ = wInUTF8.Close()

	return []CaseValidation{
		{
			Number: 0,
			HotelsInfo: HotelsInfo{
				HotelName: "Hotel Moscow",
				Stars:     -1,
				URI:       "https://test.com/path",
			},
			Expected: false,
		},
		{
			Number: 1,
			HotelsInfo: HotelsInfo{
				HotelName: "Hotel Madrid",
				Stars:     6,
				URI:       "https://test.com/path",
			},
			Expected: false,
		},
		{
			Number: 2,
			HotelsInfo: HotelsInfo{
				HotelName: "Hotel Rome",
				Stars:     4,
				URI:       "https://test.com/path",
			},
			Expected: true,
		},
		{
			Number: 3,
			HotelsInfo: HotelsInfo{
				HotelName: b.String(),
				Stars:     4,
				URI:       "https://test.com/path",
			},
			Expected: false,
		},
		{
			Number: 4,
			HotelsInfo: HotelsInfo{
				HotelName: "Hotel Berlin",
				Stars:     1,
				URI:       "//test.com/path",
			},
			Expected: false,
		},
		{
			Number: 5,
			HotelsInfo: HotelsInfo{
				HotelName: "Hotel Amsterdam",
				Stars:     1,
				URI:       "https://",
			},
			Expected: false,
		},
		{
			Number: 6,
			HotelsInfo: HotelsInfo{
				HotelName: "Hotel New-York",
				Stars:     1,
				URI:       "",
			},
			Expected: false,
		},
	}
}

type CaseSort struct {
	Number      int
	Hotels      []ConvertedHotelInfo
	SortingMode SortingMode
	Expected    []ConvertedHotelInfo
}

func TestSort(t *testing.T) {
	assrt := assert.New(t)
	for _, c := range getSortCases() {
		assrt.Equal(c.Expected, sortHotels(c.Hotels, c.SortingMode), fmt.Sprintf("Test number: %v", c.Number))
	}
}

func getSortCases() []CaseSort {
	return []CaseSort{
		{
			Number: 0,
			Hotels: []ConvertedHotelInfo{
				{
					HotelName: "Newton",
					Stars:     4,
				},
				{
					HotelName: "maxwell",
					Stars:     4,
				},
				{
					HotelName: "Planck",
					Stars:     4,
				},
			},
			SortingMode: "names",
			Expected: []ConvertedHotelInfo{
				{
					HotelName: "maxwell",
					Stars:     4,
				},
				{
					HotelName: "Newton",
					Stars:     4,
				},
				{
					HotelName: "Planck",
					Stars:     4,
				},
			},
		},
		{
			Number: 1,
			Hotels: []ConvertedHotelInfo{
				{
					HotelName: "Newton",
					Stars:     0,
				},
				{
					HotelName: "maxwell",
					Stars:     3,
				},
				{
					HotelName: "Planck",
					Stars:     5,
				},
			},
			SortingMode: "stars",
			Expected: []ConvertedHotelInfo{
				{
					HotelName: "Planck",
					Stars:     5,
				},
				{
					HotelName: "maxwell",
					Stars:     3,
				},
				{
					HotelName: "Newton",
					Stars:     0,
				},
			},
		},
		{
			Number: 2,
			Hotels: []ConvertedHotelInfo{
				{
					HotelName: "Newton",
					Stars:     0,
				},
				{
					HotelName: "maxwell",
					Stars:     2,
				},
				{
					HotelName: "Planck",
					Stars:     4,
				},
			},
			Expected: []ConvertedHotelInfo{
				{
					HotelName: "Newton",
					Stars:     0,
				},
				{
					HotelName: "maxwell",
					Stars:     2,
				},
				{
					HotelName: "Planck",
					Stars:     4,
				},
			},
		},
	}
}

func TestReadCSV(t *testing.T) {
	assrt := assert.New(t)
	for _, c := range getReadCsvCases() {
		hotels, _, _ := readCSV(bytes.NewReader(c.Bytes), c.Delimiter)
		assrt.Equal(c.Expected, hotels, fmt.Sprintf("Test number: %v", c.Number))
	}
}

type CaseReadCsv struct {
	Number    int
	Delimiter rune
	Bytes     []byte
	Expected  []ConvertedHotelInfo
}

var csvBytesComma = []byte(`
name,address,stars,contact,phone,uri
The Gibson,"63847 Lowe Knoll, East Maxine, WA 97030-4876",5,Dr. Sinda Wyman,1-270-665-9933x1626,http://www.paucek.com/search.htm
Martini Cattaneo,"Stretto Bernardi 004, Quarto Mietta nell'emilia, 07958 Torino (OG)",5,Rosalino Marchetti,+39 627 68225719,http://www.farina.org/blog/categories/tags/about.html
`)

var csvBytesDotComma = []byte(`
name;address;stars;contact;phone;uri
The Gibson;"63847 Lowe Knoll, East Maxine, WA 97030-4876";5;Dr. Sinda Wyman;1-270-665-9933x1626;http://www.paucek.com/search.htm
`)

func getReadCsvCases() []CaseReadCsv {
	return []CaseReadCsv{
		{
			Number:    0,
			Delimiter: ',',
			Bytes:     csvBytesComma,
			Expected: []ConvertedHotelInfo{
				{
					HotelName: "The Gibson",
					Address:   "63847 Lowe Knoll, East Maxine, WA 97030-4876",
					Stars:     5,
					Contact:   "Dr. Sinda Wyman",
					Phone:     "1-270-665-9933x1626",
					URI:       "http://www.paucek.com/search.htm",
				},
				{
					HotelName: "Martini Cattaneo",
					Address:   "Stretto Bernardi 004, Quarto Mietta nell'emilia, 07958 Torino (OG)",
					Stars:     5,
					Contact:   "Rosalino Marchetti",
					Phone:     "+39 627 68225719",
					URI:       "http://www.farina.org/blog/categories/tags/about.html",
				},
			},
		},
		{
			Number:    1,
			Delimiter: ';',
			Bytes:     csvBytesDotComma,
			Expected: []ConvertedHotelInfo{
				{
					HotelName: "The Gibson",
					Address:   "63847 Lowe Knoll, East Maxine, WA 97030-4876",
					Stars:     5,
					Contact:   "Dr. Sinda Wyman",
					Phone:     "1-270-665-9933x1626",
					URI:       "http://www.paucek.com/search.htm",
				},
			},
		},
	}
}
