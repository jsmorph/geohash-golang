// geohash.go
// Geohash library for Golang
// Ported from David Troy's geohash-js library (https://github.com/davetroy/geohash-js)
// (c) Tomi Hiltunen 2014
// Distributed under the MIT License

package geohash

import (
	"bytes"
	"strings"
)

var (
	bits      = []int{16, 8, 4, 2, 1}
	base32    = []byte("0123456789bcdefghjkmnpqrstuvwxyz")
	neighbors = [][]string{
		[]string{
			"p0r21436x8zb9dcf5h7kjnmqesgutwvy",
			"bc01fg45238967deuvhjyznpkmstqrwx",
		},
		[]string{
			"bc01fg45238967deuvhjyznpkmstqrwx",
			"p0r21436x8zb9dcf5h7kjnmqesgutwvy",
		},
		[]string{
			"14365h7k9dcfesgujnmqp0r2twvyx8zb",
			"238967debc01fg45kmstqrwxuvhjyznp",
		},
		[]string{
			"238967debc01fg45kmstqrwxuvhjyznp",
			"14365h7k9dcfesgujnmqp0r2twvyx8zb",
		},
	}
	borders = [][]string{
		[]string{
			"prxz",
			"bcfguvyz",
		},
		[]string{
			"bcfguvyz",
			"prxz",
		},
		[]string{
			"028b",
			"0145hjnp",
		},
		[]string{
			"0145hjnp",
			"028b",
		},
	}
)

type Direction int

const (
	North Direction = iota
	East  Direction = iota
	South Direction = iota
	West  Direction = iota
)

// Calculates adjacent geohashes.
func CalculateAdjacent(s string, dir Direction) string {
	n := len(s)
	s = strings.ToLower(s)
	lastChr := s[(len(s) - 1):]
	oddEven := (n % 2) // 0=even; 1=odd;
	// Bug fix From github.com/DeanThompson/geohash-golang
	base := s[:(n - 1)]
	if strings.Index(borders[dir][oddEven], lastChr) != -1 {
		base = CalculateAdjacent(base, dir)
	}
	return base + string(base32[strings.Index(neighbors[dir][oddEven], lastChr)])
}

// Struct for passing Box.
type BoundingBox struct {
	sw     LatLng
	ne     LatLng
	center LatLng
}

// Returns coordinates for box's center.
func (b *BoundingBox) Center() *LatLng {
	return &b.center
}

// Returns coordinates for box's South-West corner.
func (b *BoundingBox) SouthWest() *LatLng {
	return &b.sw
}

// Returns coordinates for box's North-East corner.
func (b *BoundingBox) NorthEast() *LatLng {
	return &b.ne
}

// Struct for passing LatLng values.
type LatLng struct {
	lat float64
	lng float64
}

// Returns latitude.
func (ll *LatLng) Lat() float64 {
	return ll.lat
}

// Returns longitude.
func (ll *LatLng) Lng() float64 {
	return ll.lng
}

func refineInterval(interval []float64, cd, mask int) []float64 {
	if cd&mask > 0 {
		interval[0] = (interval[0] + interval[1]) / 2
	} else {
		interval[1] = (interval[0] + interval[1]) / 2
	}
	return interval
}

// Get LatLng coordinates from a geohash
func Decode(geohash string) *BoundingBox {
	isEven := true
	lat := []float64{-90, 90}
	lng := []float64{-180, 180}
	latErr := float64(90)
	lngErr := float64(180)
	var c string
	var cd int
	for i := 0; i < len(geohash); i++ {
		c = geohash[i : i+1]
		cd = bytes.Index(base32, []byte(c))
		for j := 0; j < 5; j++ {
			if isEven {
				lngErr /= 2
				lng = refineInterval(lng, cd, bits[j])
			} else {
				latErr /= 2
				lat = refineInterval(lat, cd, bits[j])
			}
			isEven = !isEven
		}
	}
	return &BoundingBox{
		sw: LatLng{
			lat: lat[0],
			lng: lng[0],
		},
		ne: LatLng{
			lat: lat[1],
			lng: lng[1],
		},
		center: LatLng{
			lat: (lat[0] + lat[1]) / 2,
			lng: (lng[0] + lng[1]) / 2,
		},
	}
}

// Create a geohash based on LatLng coordinates
func Encode(latitude, longitude float64) string {
	isEven := true
	lat := []float64{-90, 90}
	lng := []float64{-180, 180}
	bit := 0
	ch := 0
	precision := 12
	var geohash bytes.Buffer
	var mid float64
	for geohash.Len() < precision {
		if isEven {
			mid = (lng[0] + lng[1]) / 2
			if longitude > mid {
				ch |= bits[bit]
				lng[0] = mid
			} else {
				lng[1] = mid
			}
		} else {
			mid = (lat[0] + lat[1]) / 2
			if latitude > mid {
				ch |= bits[bit]
				lat[0] = mid
			} else {
				lat[1] = mid
			}
		}
		isEven = !isEven
		if bit < 4 {
			bit++
		} else {
			geohash.WriteByte(base32[ch])
			bit = 0
			ch = 0
		}
	}
	return geohash.String()
}
