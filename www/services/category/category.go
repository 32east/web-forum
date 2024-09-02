package category

import (
	"database/sql"
	"log"
)

func GetForums(db *sql.DB) (*[]interface{}, error) {
	var forums []interface{}
	rows, err := db.Query("SELECT * FROM `forums`;")

	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var id int
		var forumName string
		var forumDescription string

		scanErr := rows.Scan(&id, &forumName, &forumDescription)

		if scanErr != nil {
			log.Fatal(scanErr)
		}

		forums = append(forums, map[string]interface{}{
			"forum_id":          id,
			"forum_name":        forumName,
			"forum_description": forumDescription,
		})
	}

	return &forums, nil
}
