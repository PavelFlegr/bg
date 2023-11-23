package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var dbName = "db"

func ensureDb() {
	_, err := os.Stat(dbName)
	if err != nil {
		f, _ := os.Create(dbName)
		f.Close()
	}
}

func readData() []Record {
	f, _ := os.Open(dbName)
	r := bufio.NewReader(f)
	data := []Record{}
	for {
		val, err := r.ReadByte()
		if err != nil {
			break
		}
		date, _ := r.ReadString('\n')
		date, _ = strings.CutSuffix(date, "\n")
		record := Record{
			Date: date,
			Data: [4]bool{
				val&1 > 0,
				val&2 > 0,
				val&4 > 0,
				val&8 > 0,
			},
		}
		data = append(data, record)
	}
	f.Close()
	return data
}

func writeData(data []Record) {
	w, _ := os.Create(dbName)
	for i := range data {
		if i != 0 {
			fmt.Fprint(w, "\n")
		}
		var val byte = 0
		for j := range data[i].Data {
			if data[i].Data[j] {
				val |= 1 << j
			}
		}
		w.Write([]byte{val})
		fmt.Fprintf(w, "%v", data[i].Date)
	}
	w.Close()
}

type Record struct {
	Date string
	Data [4]bool
}

func setValue(date string, i int, val bool, records *[]Record) {
	record := getRecord(date, *records)
	if record == nil {
		*records = append(*records, Record{
			Date: date,
			Data: [4]bool{true, true, true, true},
		})
		record = &(*records)[len(*records)-1]
	}

	record.Data[i] = val
}

func getRecord(date string, records []Record) *Record {
	for j := range records {
		if records[j].Date == date {
			return &records[j]
		}
	}

	return nil
}

func calculateShame(records []Record) [4]int {
	shame := [4]int{}
	for i := range records {
		for j := range shame {
			if !records[i].Data[j] {
				shame[j] += 1
			}
		}
	}

	return shame
}

func main() {
	ensureDb()
	records := readData()
	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		date := r.PostFormValue("date")
		i, _ := strconv.Atoi(r.PostFormValue("i"))
		val := r.PostFormValue("val") == "true"

		setValue(date, i, val, &records)
		writeData(records)
		fmt.Fprintf(w, "%v", records)
	})

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%v", records)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, _ := template.New("index.tmpl").ParseFiles("index.tmpl")
		shame := calculateShame(records)
		table := [][]string{
			{"", "Shame Score"},
			{"Blumbo", fmt.Sprint(shame[0])},
			{"Ploplop", fmt.Sprint(shame[1])},
			{"Shadowheart", fmt.Sprint(shame[2])},
			{"Xyl", fmt.Sprint(shame[3])},
		}

		date := time.Now()
		for i := 0; i < 7; i++ {
			formattedDate := date.Format("2.1.2006")
			record := getRecord(formattedDate, records)
			if record == nil {
				record = &Record{
					Date: formattedDate,
					Data: [4]bool{true, true, true, true},
				}
			}
			table[0] = append(table[0], formattedDate)
			for j := 0; j < 4; j++ {
				var checked string
				if record.Data[j] {
					checked = "checked"
				}
				table[j+1] = append(table[j+1], fmt.Sprintf("<input autocomplete='off' onclick='toggle(this, \"%v\", %v)' type='checkbox' %v>", formattedDate, j, checked))
			}
			date = date.Add(time.Hour * 24)
		}
		tmpl.Execute(w, table)
	})

	http.ListenAndServe(":8083", nil)
}
