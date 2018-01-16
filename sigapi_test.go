package main

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"math/rand"
	h "net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestSDB(t *testing.T) {
	stCols, stTable, rowAmnt := []string{
		"id_student",
		"identification",
		"name",
		"middle_name",
		"last_name",
		"address",
		"phone",
	},
		"student",
		uint32(rand.Intn(1000))

	qr := &tQr{
		t:    t,
		cnt:  make(map[string]map[string][]string),
		lcol: uint64(rowAmnt),
	}
	qr.cnt[stTable] = make(map[string][]string)
	for _, j := range stCols {
		qr.cnt[stTable][j] = randSlc(rowAmnt)
	}
	// { tQr student table filled with random strings }
	pages := uint32(rand.Intn(100))
	lim, rm, db := rowAmnt/pages, rowAmnt%pages, &SDB{Db: qr}
	if rm != 0 {
		pages = pages + 1
	}
	for i := uint32(0); i != pages; i++ {
		// { when i = pages+1, i*lim > rowAmnt,
		//   therefore i*lim is out of range and
		//   w.Code = h.StatusBadRequest }
		ul := fmt.Sprintf("http://api.com/?offset=%d&size=%d",
			i*lim, lim)
		ur, e := url.Parse(ul)
		require.NoError(t, e)
		require.True(t,
			ur.Query().Get("offset") ==
				strconv.FormatUint(uint64(i*lim), 10) &&
				ur.Query().Get("size") ==
					strconv.FormatUint(uint64(lim), 10))
		r, e := h.NewRequest(h.MethodGet, ul, nil)
		require.NoError(t, e)
		w := httptest.NewRecorder()
		db.ServeHTTP(w, r)
		require.True(t,
			(w.Code != h.StatusBadRequest && i <= pages) ||
				(w.Code == h.StatusOK || i > pages),
			"code = %d, i = %d, pages = %d, msg = %s",
			w.Code, i, pages, w.Body.String())
		if w.Code == h.StatusOK {
			sl := make([]DBRecord, 0)
			e = Decode(w.Body, &sl)
			qrsl := rgSlices(qr.cnt[stTable], stCols, i*lim, lim)
			require.True(t, equalDS(sl, qrsl),
				"%v ≠ %v At %d of %d with %d",
				qrsl, sl, i, pages, lim)
		}
	}
	require.True(t, true)
}

func rgSlices(m map[string][]string, s []string, p,
	l uint32) (d []DBRecord) {
	d = make([]DBRecord, l)
	for i := uint32(0); i != l; i++ {
		d[i] = DBRecord{}
		var name, middle_name, last_name string
		row := i + p
		for _, j := range s {
			if j == "id_student" {
				d[i].Id = m[j][row]
			} else if j == "identification" {
				d[i].IN = m[j][row]
			} else if j == "name" {
				name = m[j][row]
			} else if j == "middle_name" {
				middle_name = m[j][row]
			} else if j == "last_name" {
				last_name = m[j][row]
			} else if j == "address" {
				d[i].Addr = m[j][row]
			} else if j == "phone" {
				d[i].Tel = m[j][row]
			}
		}
		d[i].Name = name + " " + middle_name + " " + last_name
	}
	return
}

func equalDS(a, b []DBRecord) (y bool) {
	y = len(a) == len(b)
	for i := 0; y && i != len(a); i++ {
		y = a[i].Addr == b[i].Addr && a[i].IN == b[i].IN &&
			a[i].Id == b[i].Id && a[i].Name == b[i].Name &&
			a[i].Tel == b[i].Tel
	}
	return
}

// Decode decodes an io.Reader with a JSON formatted object
func Decode(r io.Reader, v interface{}) (e error) {
	var bs []byte
	bs, e = ioutil.ReadAll(r)
	if e == nil {
		e = json.Unmarshal(bs, v)
	}
	return
}

func randSlc(n uint32) (s []string) {
	s = make([]string, n)
	for i := range s {
		s[i] = strconv.FormatUint(uint64(rand.Intn(100)), 16)
	}
	return
}

type tQr struct {
	t *testing.T
	// set of tables
	cnt  map[string]map[string][]string
	lcol uint64
}

func (q *tQr) Query(s string,
	args ...interface{}) (r Scanner, e error) {
	rg := regexp.MustCompile(
		"SELECT ([[:alpha:]_,]+) FROM ([[:alpha:]_,]+)" +
			" LIMIT ([[:digit:]]+) OFFSET ([[:digit:]]+)")
	xs := rg.FindStringSubmatch(s)
	require.True(q.t, len(xs) == 5, "%d ≠ 5", len(xs))
	sFields, table, sLimit, sOffset := xs[1], xs[2], xs[3], xs[4]
	fields := strings.Split(sFields, ",")
	var limit, offset uint64
	limit, e = strconv.ParseUint(sLimit, 10, 64)
	if e == nil {
		offset, e = strconv.ParseUint(sOffset, 10, 64)
	}
	if e == nil {
		r = &tSc{
			rs:    q.cnt[table],
			frs:   fields,
			i:     offset,
			limit: limit,
			lcol:  q.lcol,
		}
	}
	return
}

type tSc struct {
	// set of columns
	rs map[string][]string
	// set of query columns
	frs []string
	// current row
	i     uint64
	limit uint64
	lcol  uint64
}

func (s *tSc) Next() (b bool) {
	b = s.i < s.lcol && s.limit != 0
	return
}

func (s *tSc) Scan(args ...interface{}) (e error) {
	if len(args) != len(s.frs) {
		e = errArgsDiffFields(len(args), len(s.frs))
	} else if !s.Next() {
		e = errNoItemsLeft()
	}
	if e == nil {
		for i, j := range args {
			v := j.(*string)
			*v = s.rs[s.frs[i]][s.i]
		}
		s.i, s.limit = s.i+1, s.limit-1
	}
	return
}

func errNoItemsLeft() (e error) {
	e = fmt.Errorf("No items left")
	return
}

func errArgsDiffFields(la, lf int) (e error) {
	e = fmt.Errorf("Len args is %d ≠ %d", la, lf)
	return
}

func TestNS(t *testing.T) {
	sdb := SDB{Db: new(tQr)}
	r, e := h.NewRequest(h.MethodPost, "http://localhost", nil)
	require.NoError(t, e)
	w := httptest.NewRecorder()
	sdb.ServeHTTP(w, r)
	require.True(t, w.Code == h.StatusBadRequest)
	require.True(t, w.Body.String() ==
		NotSuppMeth(h.MethodPost).Error())
}
