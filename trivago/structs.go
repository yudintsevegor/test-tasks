package main

// name,address,stars,contact,phone,uri
type HotelsInfo struct {
	HotelName string `csv:"name"`
	Address   string `csv:"address"`
	Stars     int    `csv:"stars"`
	Contact   string `csv:"contact"`
	Phone     string `csv:"phone"`
	URI       string `csv:"uri"`
}

type ConvertedHotelInfo struct {
	HotelName string `xml:"name" json:"name"`
	Address   string `xml:"address"  json:"address"`
	Stars     int    `xml:"stars"  json:"stars"`
	Contact   string `xml:"contact"  json:"contact"`
	Phone     string `xml:"phone"  json:"phone"`
	URI       string `xml:"uri"  json:"uri"`
}

type FormatTag string

const (
	xmlTag  FormatTag = "xml"
	jsonTag FormatTag = "json"
	csvTag  FormatTag = "csv"
)

func (tag FormatTag) String() string {
	return string(tag)
}

type SortingMode string

const (
	sortByStars SortingMode = "stars"
	sortByNames SortingMode = "names"
)
