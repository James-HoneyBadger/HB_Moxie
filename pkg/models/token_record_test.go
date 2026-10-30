package models

import (
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/James-HoneyBadger/honey-ledger/pkg/utils"
)

func TestTokenRecordSecretColumnFitsGeneratedTokenSecret(t *testing.T) {
	secret, err := utils.GetRandomString(32)
	assert.NoError(t, err)

	field, exists := reflect.TypeOf(TokenRecord{}).FieldByName("Secret")
	assert.True(t, exists)
	assert.Contains(t, field.Tag.Get("xorm"), "VARCHAR(64)")
	assert.LessOrEqual(t, len(secret), 64)
}

func TestTokenInfoResponseSliceLess(t *testing.T) {
	var tokenInfoRespSlice TokenInfoResponseSlice
	tokenInfoRespSlice = append(tokenInfoRespSlice, &TokenInfoResponse{
		TokenId:  "1",
		LastSeen: int64(1),
	})
	tokenInfoRespSlice = append(tokenInfoRespSlice, &TokenInfoResponse{
		TokenId:  "2",
		LastSeen: int64(3),
	})
	tokenInfoRespSlice = append(tokenInfoRespSlice, &TokenInfoResponse{
		TokenId:  "3",
		LastSeen: int64(2),
	})

	sort.Sort(tokenInfoRespSlice)

	assert.Equal(t, "2", tokenInfoRespSlice[0].TokenId)
	assert.Equal(t, "3", tokenInfoRespSlice[1].TokenId)
	assert.Equal(t, "1", tokenInfoRespSlice[2].TokenId)
}
