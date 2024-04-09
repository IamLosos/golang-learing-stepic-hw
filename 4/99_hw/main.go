package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

// код писать тут

type PersonsDataset struct {
	Rows []Row `xml:"row"`
}

type Row struct {
	Id            int    `xml:"id"`
	Guid          string `xml:"guid"`
	IsActive      string `xml:"isActive"`
	Balance       string `xml:"balance"`
	Picture       string `xml:"picture"`
	Age           int    `xml:"age"`
	EyeColor      string `xml:"eyeColor"`
	First_name    string `xml:"first_name"`
	Last_name     string `xml:"last_name"`
	Gender        string `xml:"gender"`
	Company       string `xml:"company"`
	Email         string `xml:"email"`
	Phone         string `xml:"phone"`
	Address       string `xml:"address"`
	About         string `xml:"about"`
	Registered    string `xml:"registered"`
	FavoriteFruit string `xml:"favoriteFruit"`
}

func (r *Row) Name() string {
	return r.First_name + r.Last_name
}

func (r Row) String() string {
	var sb strings.Builder
	sb.WriteString("*")
	sb.WriteString(strconv.Itoa(r.Id))
	sb.WriteString(" ")
	sb.WriteString(r.Name())
	sb.WriteString(" ")
	sb.WriteString(strconv.Itoa(r.Age))
	sb.WriteString("*")
	//sb.WriteString(r.About)

	return sb.String()
}

func Filter[T any](s []T, cond func(t T) bool) []T {
	res := make([]T, 40)
	i := 0
	for _, v := range s {
		if cond(v) {
			res[i] = v
			i++
		}
	}

	return res[0:i]
}

func (p *PersonsDataset) PrintLen(prefix string) {
	if p.Rows == nil {
		fmt.Printf("%s:rows is nil\r\n", prefix)
	} else {
		fmt.Printf("%s:len %d\r\n", prefix, len(p.Rows))
	}
}

func (p *PersonsDataset) GetCopy() PersonsDataset {
	cp := PersonsDataset{}
	//copy(cp.Rows, p.Rows)
	cp.Rows = p.Rows
	return cp
}

func (p *PersonsDataset) Init() {
	body, err := ioutil.ReadFile("dataset.xml")
	if err != nil {
		panic(err)
	}

	err = xml.Unmarshal(body, p)
	if err != nil {
		panic(err)
	}
}

func (p *PersonsDataset) Filter(query string) {
	fr := Filter(p.Rows, func(r Row) bool {
		return query == "" || strings.Contains(r.Name(), query) || strings.Contains(r.About, query)
	})

	p.Rows = fr
}

func (p *PersonsDataset) Sort(orderField string, orderBy int) {
	switch orderBy {
	case OrderByAsc, OrderByDesc:
		var f func(i, j int) bool
		switch orderField {
		case "Id":
			f = func(i, j int) bool {
				return orderBy == OrderByAsc && p.Rows[i].Id < p.Rows[j].Id || orderBy == OrderByDesc && p.Rows[i].Id > p.Rows[j].Id
			}
		case "Age":
			f = func(i, j int) bool {
				return orderBy == OrderByAsc && p.Rows[i].Age < p.Rows[j].Age || orderBy == OrderByDesc && p.Rows[i].Age > p.Rows[j].Age
			}
		case "Name", "":
			f = func(i, j int) bool {
				return orderBy == OrderByAsc && strings.Compare(p.Rows[i].Name(), p.Rows[j].Name()) == -1 || orderBy == OrderByDesc && strings.Compare(p.Rows[i].Name(), p.Rows[j].Name()) == 1
			}
		default:
			panic("Wrong orderField value: " + orderField)
		}

		sort.Slice(p.Rows, f)
	case OrderByAsIs:
		break
	default:
		panic(fmt.Sprintf("Wrong orderBy value: %d", orderBy))
	}
}

func (p *PersonsDataset) Cut(limit int, offset int) {
	if offset > len(p.Rows) {
		p.Rows = make([]Row, 0)
		return
	}

	e := offset + limit
	if e > len(p.Rows) {
		e = len(p.Rows)
	}

	p.Rows = p.Rows[offset:e]
}

var basicPersonsDataset PersonsDataset

func (p *PersonsDataset) SearchServer(query string, orderField string, orderBy int, limit int, offset int) []Row {
	p.PrintLen("p.SearchServer:p")
	p2 := p.GetCopy()
	p2.PrintLen("p.SearchServer:p2")

	p2.Filter(query)
	p2.Sort(orderField, orderBy)
	p2.Cut(limit, offset)

	p2.PrintLen("p.SearchServer:p2 after search")
	return p2.Rows
}

func (p *PersonsDataset) SearchServerByQuerryParams(qp QueryParams) []Row {
	return p.SearchServer(qp.Querry, qp.OrderField, qp.OrderBy, qp.Limit, qp.Offset)
}

type QueryParams struct {
	Querry     string
	OrderField string
	OrderBy    int
	Limit      int
	Offset     int
}

func (p *QueryParams) CheckParams() error {
	if !slices.Contains([]string{"Id", "Age", "Name", ""}, p.OrderField) {
		return fmt.Errorf("'order_field' wrong value %s", p.OrderField)
	}

	if !slices.Contains([]int{OrderByAsIs, OrderByAsc, OrderByDesc}, p.OrderBy) {
		return fmt.Errorf("'order_by' wrong value %d", p.OrderBy)
	}

	if p.Limit <= 0 {
		return fmt.Errorf("'limit' wrong value %d", p.Limit)
	}

	if p.Offset < 0 {
		return fmt.Errorf("'offset' wrong value %d", p.Offset)
	}

	return nil
}

func (p *QueryParams) ParseParams(r *http.Request) error {
	p.Querry = r.URL.Query().Get("query")
	p.OrderField = r.URL.Query().Get("order_field")

	var err error
	p.OrderBy, err = strconv.Atoi(r.URL.Query().Get("order_by"))
	if err != nil {
		return fmt.Errorf("error %s for field '%s'", err, "order_by")
	}

	p.Limit, err = strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		return fmt.Errorf("error %s for field '%s'", err, "limit")
	}

	p.Offset, err = strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		return fmt.Errorf("error %s for field '%s'", err, "offset")
	}

	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	var qp QueryParams = QueryParams{}
	err := qp.ParseParams(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	err = qp.CheckParams()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	for _, r := range basicPersonsDataset.SearchServerByQuerryParams(qp) {
		w.Write([]byte(fmt.Sprintln(r)))
	}
}

func runServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	server := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("starting server at :8080")
	server.ListenAndServe()

}

func main() {
	basicPersonsDataset = PersonsDataset{}
	basicPersonsDataset.Init()
	// basicPersonsDataset.PrintLen("main:basicPersonsDataset")

	runServer()
}

func consoleTests() {
	var p []Row

	//fmt.Println("An - OrderByAsIs - Id")
	//p = basicPersonsDataset.SearchServer("A", "Id", OrderByAsc, 22, 0)
	//p = SearchServer("A", "Id", OrderByAsc, 22, 0)
	//fmt.Println(p)
	//return

	fmt.Println("An - OrderByAsIs - Id")
	p = basicPersonsDataset.SearchServer("An", "Id", OrderByAsIs, 1, 1)
	fmt.Println(p)

	fmt.Println("An - OrderByAsc - Id")
	p = basicPersonsDataset.SearchServer("An", "Id", OrderByAsc, 100, 0)
	fmt.Println(p)

	fmt.Println("An - OrderByDesc - Id")
	p = basicPersonsDataset.SearchServer("An", "Id", OrderByDesc, 100, 0)
	fmt.Println(p)

	//
	fmt.Println("---")
	//

	fmt.Println("An - OrderByAsIs - Name")
	p = basicPersonsDataset.SearchServer("An", "Name", OrderByAsIs, 100, 0)
	fmt.Println(p)

	fmt.Println("An - OrderByAsc - Name")
	p = basicPersonsDataset.SearchServer("An", "Name", OrderByAsc, 100, 0)
	fmt.Println(p)

	fmt.Println("An - OrderByDesc - Name")
	p = basicPersonsDataset.SearchServer("An", "Name", OrderByDesc, 100, 0)
	fmt.Println(p)

	//
	fmt.Println("---")
	//

	fmt.Println("An - OrderByAsIs - ''")
	p = basicPersonsDataset.SearchServer("An", "", OrderByAsIs, 100, 0)
	fmt.Println(p)

	fmt.Println("An - OrderByAsc - ''")
	p = basicPersonsDataset.SearchServer("An", "", OrderByAsc, 100, 0)
	fmt.Println(p)

	fmt.Println("An - OrderByDesc - ''")
	p = basicPersonsDataset.SearchServer("An", "", OrderByDesc, 100, 0)
	fmt.Println(p)

	//
	fmt.Println("---")
	//

	fmt.Println("An - OrderByAsIs - Age")
	p = basicPersonsDataset.SearchServer("An", "Age", OrderByAsIs, 100, 0)
	fmt.Println(p)

	fmt.Println("An - OrderByAsc - Age")
	p = basicPersonsDataset.SearchServer("An", "Age", OrderByAsc, 100, 0)
	fmt.Println(p)

	fmt.Println("An - OrderByDesc - Age")
	p = basicPersonsDataset.SearchServer("An", "Age", OrderByDesc, 100, 0)
	fmt.Println(p)
}
