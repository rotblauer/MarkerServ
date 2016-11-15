package markerMaker

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

func parseForm(r *http.Request, formVal string) []string {
	return strings.Split(r.FormValue(formVal), "\n")
}

func parseBedForm(r *http.Request, formVal string) []Bed3 {
	return nil
}

const MaxUint = ^uint(0)
const MaxInt = int(MaxUint >> 1)

const (
	chromField = iota
	startField
	endField
)

type Bed3 struct {
	Chrom      string
	ChromStart int
	ChromEnd   int
}

func parseBed3(line string) (b *Bed3, err error) {
	const n = 1
	tmp := strings.Replace(line, "chr", "", -1)
	tmp2 := strings.Replace(tmp, ",", "", -1)

	f := strings.FieldsFunc(tmp2, func(r rune) bool {
		return r == ':' || r == '-'
	})
	if len(f) < 1 {
		return nil, errors.New("bed: bad bed type")
	}
	chr := string(f[chromField])
	var start int
	var stop int
	if len(f) == 1 {
		start = 0
		stop = MaxInt
	} else if len(f) == 2 {
		start, err = mustAtoi(f[startField], startField)
		if err != nil {
			return nil, errors.New("bed: bad bed type " + f[startField])
		}
		stop = start
	} else {
		start, err = mustAtoi(f[startField], startField)
		if err != nil {
			return nil, errors.New("bed: bad bed type " + f[startField])
		}
		stop, err = mustAtoi(f[endField], endField)
		if err != nil {
			return nil, errors.New("bed: bad bed type " + f[endField])
		}
	}
	b = &Bed3{
		Chrom:      chr,
		ChromStart: start,
		ChromEnd:   stop,
	}
	return b, nil
}
func mustAtoi(f string, column int) (int, error) {
	i, err := strconv.ParseInt(f, 0, 0)
	if err != nil {
		return -1, err
	}
	return int(i), nil
}

func (b *Bed3) Start() int { return b.ChromStart }
func (b *Bed3) End() int   { return b.ChromEnd }
func (b *Bed3) Len() int   { return b.ChromEnd - b.ChromStart }
