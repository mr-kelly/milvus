// Licensed to the LF AI & Data foundation under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"testing"

	"github.com/milvus-io/milvus/internal/util/funcutil"
	"github.com/milvus-io/milvus/internal/util/uniquegenerator"
	"github.com/stretchr/testify/assert"
)

func TestIndexFileBinlogCodec(t *testing.T) {
	indexBuildID := UniqueID(uniquegenerator.GetUniqueIntGeneratorIns().GetInt())
	version := int64(uniquegenerator.GetUniqueIntGeneratorIns().GetInt())
	collectionID := UniqueID(uniquegenerator.GetUniqueIntGeneratorIns().GetInt())
	partitionID := UniqueID(uniquegenerator.GetUniqueIntGeneratorIns().GetInt())
	segmentID := UniqueID(uniquegenerator.GetUniqueIntGeneratorIns().GetInt())
	fieldID := UniqueID(uniquegenerator.GetUniqueIntGeneratorIns().GetInt())
	indexName := funcutil.GenRandomStr()
	indexID := UniqueID(uniquegenerator.GetUniqueIntGeneratorIns().GetInt())
	indexParams := make(map[string]string)
	indexParams["index_type"] = "IVF_FLAT"
	datas := []*Blob{
		{
			Key:   "ivf1",
			Value: []byte{1, 2, 3},
		},
		{
			Key:   "ivf2",
			Value: []byte{4, 5, 6},
		},
		{
			Key:   "large",
			Value: funcutil.RandomBytes(maxLengthPerRowOfIndexFile + 1),
		},
	}

	codec := NewIndexFileBinlogCodec()

	serializedBlobs, err := codec.Serialize(indexBuildID, version, collectionID, partitionID, segmentID, fieldID, indexParams, indexName, indexID, datas)
	assert.Nil(t, err)

	idxBuildID, v, collID, parID, segID, fID, params, idxName, idxID, blobs, err := codec.DeserializeImpl(serializedBlobs)
	assert.Nil(t, err)
	assert.Equal(t, indexBuildID, idxBuildID)
	assert.Equal(t, version, v)
	assert.Equal(t, collectionID, collID)
	assert.Equal(t, partitionID, parID)
	assert.Equal(t, segmentID, segID)
	assert.Equal(t, fieldID, fID)
	assert.Equal(t, len(indexParams), len(params))
	for key, value := range indexParams {
		assert.Equal(t, value, params[key])
	}
	assert.Equal(t, indexName, idxName)
	assert.Equal(t, indexID, idxID)
	assert.ElementsMatch(t, datas, blobs)

	blobs, indexParams, indexName, indexID, err = codec.Deserialize(serializedBlobs)
	assert.Nil(t, err)
	assert.ElementsMatch(t, datas, blobs)
	for key, value := range indexParams {
		assert.Equal(t, value, params[key])
	}
	assert.Equal(t, indexName, idxName)
	assert.Equal(t, indexID, idxID)

	// empty
	_, _, _, _, _, _, _, _, _, _, err = codec.DeserializeImpl(nil)
	assert.NotNil(t, err)
}

func TestIndexFileBinlogCodecError(t *testing.T) {
	var err error

	// failed to read binlog
	codec := NewIndexFileBinlogCodec()
	_, _, _, _, err = codec.Deserialize([]*Blob{{Key: "key", Value: []byte("not in binlog format")}})
	assert.NotNil(t, err)

	indexBuildID := UniqueID(uniquegenerator.GetUniqueIntGeneratorIns().GetInt())
	version := int64(uniquegenerator.GetUniqueIntGeneratorIns().GetInt())
	collectionID := UniqueID(uniquegenerator.GetUniqueIntGeneratorIns().GetInt())
	partitionID := UniqueID(uniquegenerator.GetUniqueIntGeneratorIns().GetInt())
	segmentID := UniqueID(uniquegenerator.GetUniqueIntGeneratorIns().GetInt())
	fieldID := UniqueID(uniquegenerator.GetUniqueIntGeneratorIns().GetInt())
	indexName := funcutil.GenRandomStr()
	indexID := UniqueID(uniquegenerator.GetUniqueIntGeneratorIns().GetInt())
	indexParams := make(map[string]string)
	indexParams["index_type"] = "IVF_FLAT"
	datas := []*Blob{
		{
			Key:   "ivf1",
			Value: []byte{1, 2, 3},
		},
	}

	_, err = codec.Serialize(indexBuildID, version, collectionID, partitionID, segmentID, fieldID, indexParams, indexName, indexID, datas)
	assert.Nil(t, err)
}

func TestIndexCodec(t *testing.T) {
	indexCodec := NewIndexCodec()
	blobs := []*Blob{
		{
			"12345",
			[]byte{1, 2, 3, 4, 5, 6, 7, 1, 2, 3, 4, 5, 6, 7},
			14,
		},
		{
			"6666",
			[]byte{6, 6, 6, 6, 6, 1, 2, 3, 4, 5, 6, 7},
			12,
		},
		{
			"8885",
			[]byte{8, 8, 8, 8, 8, 8, 8, 8, 2, 3, 4, 5, 6, 7},
			14,
		},
	}
	indexParams := map[string]string{
		"k1": "v1", "k2": "v2",
	}
	blobsInput, err := indexCodec.Serialize(blobs, indexParams, "index_test_name", 1234)
	assert.Nil(t, err)
	assert.EqualValues(t, 4, len(blobsInput))
	assert.EqualValues(t, IndexParamsKey, blobsInput[3].Key)
	blobsOutput, indexParamsOutput, indexName, indexID, err := indexCodec.Deserialize(blobsInput)
	assert.Nil(t, err)
	assert.EqualValues(t, 3, len(blobsOutput))
	for i := 0; i < 3; i++ {
		assert.EqualValues(t, blobs[i], blobsOutput[i])
	}
	assert.EqualValues(t, indexParams, indexParamsOutput)
	assert.EqualValues(t, "index_test_name", indexName)
	assert.EqualValues(t, 1234, indexID)

	blobs = []*Blob{}
	_, _, _, _, err = indexCodec.Deserialize(blobs)
	assert.NotNil(t, err)
}
