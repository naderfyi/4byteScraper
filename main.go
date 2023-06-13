package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	sigDB *gorm.DB
	err   error
)

func ensure_sig_db(db *gorm.DB) error {
	cfg := getConfig()

	host := cfg.Postgres.Host
	port := cfg.Postgres.Port
	user := cfg.Postgres.User
	password := cfg.Postgres.Password
	dbsig := cfg.Postgres.DBSig

	// Create the signature database if it doesn't exist
	var count_sig int
	err := db.Table("pg_database").Where("datname = ?", dbsig).Count(&count_sig).Error
	if err != nil {
		return err
	}
	if count_sig == 0 {
		_, err := db.DB().Exec("CREATE DATABASE " + dbsig)
		if err != nil {
			return err
		}
		color.Green("Signatures Database Created!")
		color.Yellow("Scraping Signatures from 4byte.directory...")
		color.Yellow("This may take a while...")
	}

	// Create the new database

	if err != nil {
		if !strings.Contains(err.Error(), "database exists") && !strings.Contains(err.Error(), "existiert bereits") {
			return err
		}
	}
	// Connect to the new database
	// Construct the connection string
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbsig)
	// Connect to the newly created database
	db, err = gorm.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}
	// Create tables if they don't exists
	if !db.HasTable(&Signature{}) {
		db.CreateTable(&Signature{})
	}
	if !db.HasTable(&Resume{}) {
		db.CreateTable(&Resume{})
	}

	var resume []Resume
	err = db.Find(&resume).Error
	if err != nil {
		return err
	}

	if len(resume) > 1 {
		return errors.New("More than one resume entry in database!")
	}
	var url string
	if len(resume) == 1 {
		// If the progress was marked as complete, we're done here.
		if resume[0].Url == "Completed" {
			return nil
		}
		// continue from the stored url
		url = resume[0].Url
	} else {
		// start at the first page
		url = "https://www.4byte.directory/api/v1/signatures/?page=1"
		db.Create(Resume{Url: url})
	}

	for {
		fmt.Println(url)
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		// if we get a non-200 HTTP code we just retry
		if res.StatusCode != 200 {
			fmt.Println("Got status:", res.Status)
			fmt.Println("Retrying...")
			continue
		}

		// read the response to a buffer
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		// parse the json response
		resp_obj := ApiRes{}
		err = json.Unmarshal(body, &resp_obj)
		if err != nil {
			return err
		}
		url = resp_obj.Next

		// iterate over all returned signatures
		for _, entry := range resp_obj.Results {
			// strip the "0x" prefix and decode the 4-byte selector
			binSig, err := hex.DecodeString(entry.HexSignature[2:])
			if err != nil {
				return err
			}
			// insert into db
			err = db.Create(&Signature{Code: binSig, Signature: entry.TextSignature}).Error
			if err != nil {
				return err
			}
		}

		// break if there's no next page
		if url == "" {
			// mark the scrapping as complete
			err = db.Model(&Resume{}).Update("Url", "Completed").Error
			break
		}

		err = db.Model(&Resume{}).Update("Url", url).Error
		if err != nil {
			return err
		}
	}
	color.Yellow("Finished Scrapping 4bytes.directory!")
	return nil
}

func main() {
	cfg := getConfig()

	// Store the configuration details in variables
	host := cfg.Postgres.Host
	port := cfg.Postgres.Port
	user := cfg.Postgres.User
	password := cfg.Postgres.Password

	// Construct the connection string
	psqlInfoSig := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable", host, port, user, password)
	// Connect to Postgres
	sigDB, err = gorm.Open("postgres", psqlInfoSig)
	if err != nil {
		log.Fatal(err)
	}
	defer sigDB.Close()

	// Create the signature database if it doesn't exist
	err = ensure_sig_db(sigDB)
	if err != nil {
		log.Fatal(err)
	}
	// Close the existing connection
	sigDB.Close()
}
