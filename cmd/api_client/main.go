package main

import (
	"fmt"
	"reflect"

	"github.com/hschmale16/the_final_stockbot/internal/m"
	"gorm.io/gorm"
)

func main() {
	pgDB, err := m.GetPostgresqlDB()
	if err != nil {
		panic(err)
	}
	err = m.ApplyMigrations(pgDB)
	if err != nil {
		panic(err)
	}

	sqliteDB, err := m.GetSqliteDB()
	if err != nil {
		panic(err)
	}
	err = m.ApplyMigrations(sqliteDB)
	if err != nil {
		panic(err)
	}

	CopyData[m.DB_CongressMember](pgDB, sqliteDB)
	CopyData[m.Tag](pgDB, sqliteDB)
	CopyData[m.SearchQuery](pgDB, sqliteDB)
	CopyData[m.GovtRssItemTag](pgDB, sqliteDB)
	CopyData[m.DB_CongressCommittee](pgDB, sqliteDB)
	CopyData[m.GovtRssItem](pgDB, sqliteDB)
	CopyData[m.GovtLawText](pgDB, sqliteDB)
	CopyData[m.DB_CommitteeMembership](pgDB, sqliteDB)
	CopyData[m.CongressMemberSponsored](pgDB, sqliteDB)
	CopyData[m.LawOffset](pgDB, sqliteDB)
}

func CopyData[MODEL any](psql *gorm.DB, sqlite *gorm.DB) error {
	// Create an instance of the type MODEL using reflection
	modelType := reflect.TypeOf((*MODEL)(nil)).Elem()
	modelInstance := reflect.New(modelType).Interface()

	// check counts to see if there's an issue
	var sqliteCount, pgCount int64
	sqlite.Model(modelInstance).Count(&sqliteCount)
	psql.Model(modelInstance).Count(&pgCount)

	if sqliteCount == pgCount {
		fmt.Println("Migration on this table is already complete: ", modelType, "rowCount sql/pg", sqliteCount, pgCount)
		return nil
	} else {
		fmt.Println("Migration on this table is not complete: ", modelType, "rowCount sql/pg", sqliteCount, pgCount)

	}

	var data []MODEL
	err := sqlite.Model(modelInstance).FindInBatches(&data, 10, func(tx *gorm.DB, batch int) error {
		err := psql.Create(data)
		if err.Error != nil {
			fmt.Println(err)
		}
		return nil
	})

	if err != nil {
		fmt.Println(err.Error)
	}

	sqlite.Model(modelInstance).Count(&sqliteCount)
	psql.Model(modelInstance).Count(&pgCount)

	if sqliteCount == pgCount {
		fmt.Println("Migration on this table is already complete: ", modelType)
		return nil
	} else {
		fmt.Println("Migration on this table is not complete: ", modelType, "rowCount sql/pg", sqliteCount, pgCount)

	}

	return nil
}

// func Old() {

// 	// token := os.Getenv("CONGRESS_GOV_API_TOKEN")

// 	// if token == "" {
// 	// 	panic("CONGRESS_GOV_API_TOKEN is not set")
// 	// }

// 	// client := congressgov.NewClient(token)

// 	// actions, err := client.GetBillActions(118, 4548, "s")
// 	// if err != nil {
// 	// 	panic(err)
// 	// }

// 	// for _, action := range actions.Actions {
// 	// 	fmt.Println(action.ActionDate, " - ", action.ActionText)
// 	// }

// 	// data, err := client.GetLatestBillActions()
// 	// if err != nil {
// 	// 	panic(err)
// 	// }

// 	// for _, bill := range data.Bills {
// 	// 	fmt.Println(bill.Congress, " - ", bill.OriginChamber, " - ", bill.Number)
// 	// 	fmt.Print((bill.Type), " - ", bill.Title)
// 	// }
// }
