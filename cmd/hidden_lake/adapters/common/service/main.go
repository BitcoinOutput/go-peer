package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/number571/go-peer/pkg/storage/database"
)

const (
	databasePath = "common_service.db"
	dataCountKey = "count_service"
)

var (
	mtx sync.Mutex
	db  database.IKeyValueDB
)

func initDB() database.IKeyValueDB {
	var err error
	db, err = database.NewSQLiteDB(
		database.NewSettings(&database.SSettings{
			FPath: databasePath,
		}),
	)
	if err != nil {
		panic(err)
	}
	if _, err := db.Get([]byte(dataCountKey)); err == nil {
		return db
	}
	if err := db.Set([]byte(dataCountKey), []byte("0")); err != nil {
		panic(err)
	}
	return db
}

func main() {
	db := initDB()
	defer db.Close()

	if len(os.Args) != 2 {
		panic("./service [port]")
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/size", sizePage)
	http.HandleFunc("/push", pushPage)
	http.HandleFunc("/load", loadPage)

	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func pushPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("PATH: %s; METHOD: %s;\n", r.URL.Path, r.Method)

	if r.Method != http.MethodPost {
		fmt.Fprint(w, "!incorrect method")
		return
	}

	res, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprint(w, "!read body")
		return
	}

	if err := pushDataToDB(res); err != nil {
		fmt.Fprint(w, "!push to database")
		return
	}

	fmt.Fprint(w, ".")
}

func sizePage(w http.ResponseWriter, r *http.Request) {
	log.Printf("PATH: %s; METHOD: %s;\n", r.URL.Path, r.Method)

	if r.Method != http.MethodGet {
		fmt.Fprint(w, "!incorrect method")
		return
	}

	count, err := countOfDataInDB()
	if err != nil {
		fmt.Fprint(w, "!read count of data")
		return
	}

	fmt.Fprintf(w, ".%d", count)
}

func loadPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("PATH: %s; METHOD: %s;\n", r.URL.Path, r.Method)

	if r.Method != http.MethodGet {
		fmt.Fprint(w, "!incorrect method")
		return
	}

	query := r.URL.Query()
	strDataID := query.Get("data_id")

	dataID, err := strconv.ParseUint(strDataID, 10, 64)
	if err != nil {
		fmt.Fprint(w, "!decode data_id")
		return
	}

	data, err := loadDataFromDB(dataID)
	if err != nil {
		fmt.Fprint(w, "!load data by data_id")
		return
	}

	w.Write(bytes.Join([][]byte{[]byte("."), data}, []byte{}))
}

func countOfDataInDB() (uint64, error) {
	res, err := db.Get([]byte(dataCountKey))
	if err != nil {
		return 0, err
	}

	count, err := strconv.ParseUint(string(res), 10, 64)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func loadDataFromDB(dataID uint64) ([]byte, error) {
	data, err := db.Get([]byte(fmt.Sprintf("%d", dataID)))
	if err != nil {
		return nil, err
	}

	return data, nil
}

func pushDataToDB(data []byte) error {
	mtx.Lock()
	defer mtx.Unlock()

	count, err := countOfDataInDB()
	if err != nil {
		return err
	}

	if err := db.Set([]byte(fmt.Sprintf("%d", count)), data); err != nil {
		return err
	}

	if err := db.Set([]byte(dataCountKey), []byte(fmt.Sprintf("%d", count+1))); err != nil {
		return err
	}

	return nil
}
