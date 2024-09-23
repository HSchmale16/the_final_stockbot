package votes_test

import (
	_ "embed"
	"log"
	"testing"

	"github.com/hschmale16/the_final_stockbot/internal/m"
	"github.com/hschmale16/the_final_stockbot/internal/votes"
	"github.com/stretchr/testify/assert"
)

//go:embed test_data/senate_test.xml
var senate_xml []byte

func TestLoadSenateXML(t *testing.T) {
	// We are loading our xml

	result := votes.ProcessSenateXml(senate_xml)

	log.Default().Print("Da Result", result.VoteModify)

	assert.Equal(t, 2024, result.VoteDate.Year(), "VoteDate is not correct")
	assert.Equal(t, 2024, result.VoteModify.Year(), "VoteModify is not correct")
	assert.Equal(t, 100, len(result.Members), "Members is not correct")
	assert.Equal(t, "2", result.Session, "Session is not correct")
	assert.Equal(t, 118, result.Congress, "Congress is not correct")
	assert.Equal(t, 212, result.VoteNumber, "VoteNumber is not correct")
	assert.Equal(t, 3, len(result.VoteDetails), "VoteDetails is not correct")

}

func TestDBInsert(t *testing.T) {
	db, err := m.SetupDB()

	assert.Nil(t, err)

	err = votes.ProcessSenateRollCall("test.xml", votes.ProcessSenateXml(senate_xml), db)
	assert.Nil(t, err)
}
