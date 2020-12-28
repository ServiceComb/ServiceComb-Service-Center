package sd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestListWatchConfig_String(t *testing.T) {
	t.Run("TestListWatchConfig_String", func(t *testing.T) {
		config := ListWatchConfig{
			Timeout: 666,
		}
		ret := config.String()
		assert.Equal(t, "{timeout: 666ns}", ret)
	})
	t.Run("when time is nil", func(t *testing.T) {
		config := ListWatchConfig{}
		ret := config.String()
		assert.Equal(t, "{timeout: 0s}", ret)
	})
}

func TestDoParseWatchRspToMongoInfo(t *testing.T) {
	documentID := primitive.NewObjectID()

	mockDocument, _ := bson.Marshal(bson.M{"_id": documentID, "domain": "default", "project": "default", "instanceinfo": bson.M{"instanceid": "8064a600438511eb8584fa163e8a81c9", "serviceid": "91afbe0faa9dc1594689139f099eb293b0cd048d",
		"hostname": "ecs-hcsadlab-dev-0002", "status": "UP", "timestamp": "1608552622", "modtimestamp": "1608552622", "version": "0.0.1"}})

	mockServiceDocument, _ := bson.Marshal(bson.M{"_id": documentID, "domain": "default", "project": "default", "serviceinfo": bson.M{"serviceid": "91afbe0faa9dc1594689139f099eb293b0cd048d", "timestamp": "1608552622", "modtimestamp": "1608552622", "version": "0.0.1"}})

	// case instance insertOp

	mockWatchRsp := &MongoWatchResponse{OperationType: insertOp,
		FullDocument: mockDocument,
		DocumentKey:  MongoDocument{ID: documentID},
	}
	ilw := innerListWatch{
		Key: instance,
	}
	info := ilw.doParseWatchRspToMongoInfo(mockWatchRsp)
	assert.Equal(t, documentID.Hex(), info.DocumentID)
	assert.Equal(t, "8064a600438511eb8584fa163e8a81c9", info.BusinessID)

	// case updateOp
	mockWatchRsp.OperationType = updateOp
	info = ilw.doParseWatchRspToMongoInfo(mockWatchRsp)
	assert.Equal(t, documentID.Hex(), info.DocumentID)
	assert.Equal(t, "8064a600438511eb8584fa163e8a81c9", info.BusinessID)
	assert.Equal(t, "1608552622", info.Value.(Instance).InstanceInfo.ModTimestamp)

	// case delete
	mockWatchRsp.OperationType = deleteOp
	info = ilw.doParseWatchRspToMongoInfo(mockWatchRsp)
	assert.Equal(t, documentID.Hex(), info.DocumentID)
	assert.Equal(t, "", info.BusinessID)

	// case service insertOp
	mockWatchRsp = &MongoWatchResponse{OperationType: insertOp,
		FullDocument: mockServiceDocument,
		DocumentKey:  MongoDocument{ID: primitive.NewObjectID()},
	}
	ilw.Key = service
	info = ilw.doParseWatchRspToMongoInfo(mockWatchRsp)
	assert.Equal(t, documentID.Hex(), info.DocumentID)
	assert.Equal(t, "91afbe0faa9dc1594689139f099eb293b0cd048d", info.BusinessID)
}

func TestInnerListWatch_ResumeToken(t *testing.T) {
	ilw := innerListWatch{
		Key:         instance,
		resumeToken: bson.Raw("resumToken"),
	}
	t.Run("get resume token test", func(t *testing.T) {
		res := ilw.ResumeToken()
		assert.NotNil(t, res)
		assert.Equal(t, bson.Raw("resumToken"), res)
	})

	t.Run("set resume token test", func(t *testing.T) {
		ilw.setResumeToken(bson.Raw("resumToken2"))
		assert.Equal(t, ilw.resumeToken, bson.Raw("resumToken2"))
	})
}