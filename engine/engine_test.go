package engine

import (
	"github.com/huichen/wukong/types"
	"github.com/huichen/wukong/utils"
	"reflect"
	"testing"
)

type ScoringFields struct {
	a, b, c float32
}

func AddDocs(engine *Engine) {
	docId := uint64(0)
	engine.IndexDocument(docId, types.DocumentIndexData{
		Content: "中国有十三亿人口人口",
		Fields:  ScoringFields{1, 2, 3},
	})
	docId++
	engine.IndexDocument(docId, types.DocumentIndexData{
		Content: "中国人口",
		Fields:  nil,
	})
	docId++
	engine.IndexDocument(docId, types.DocumentIndexData{
		Content: "有人口",
		Fields:  ScoringFields{2, 3, 1},
	})
	docId++
	engine.IndexDocument(docId, types.DocumentIndexData{
		Content: "有十三亿人口",
		Fields:  ScoringFields{2, 3, 3},
	})
	docId++
	engine.IndexDocument(docId, types.DocumentIndexData{
		Content: "中国十三亿人口",
		Fields:  ScoringFields{0, 9, 1},
	})

	engine.FlushIndex()
}

type RankByTokenProximity struct {
}

func (rule RankByTokenProximity) Score(
	doc types.IndexedDocument, fields interface{}) []float32 {
	if doc.TokenProximity < 0 {
		return []float32{}
	}
	return []float32{1.0 / (float32(doc.TokenProximity) + 1)}
}

func TestEngineIndexDocument(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineInitOptions{
		SegmenterDictionaries: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.LocationsIndex,
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchRequest{Text: "中国人口"})
	utils.Expect(t, "2", len(outputs.Tokens))
	utils.Expect(t, "中国", outputs.Tokens[0])
	utils.Expect(t, "人口", outputs.Tokens[1])
	utils.Expect(t, "3", len(outputs.Docs))

	utils.Expect(t, "1", outputs.Docs[0].DocId)
	utils.Expect(t, "1000", int(outputs.Docs[0].Scores[0]*1000))
	utils.Expect(t, "[0 6]", outputs.Docs[0].TokenSnippetPositions)

	utils.Expect(t, "4", outputs.Docs[1].DocId)
	utils.Expect(t, "100", int(outputs.Docs[1].Scores[0]*1000))
	utils.Expect(t, "[0 15]", outputs.Docs[1].TokenSnippetPositions)

	utils.Expect(t, "0", outputs.Docs[2].DocId)
	utils.Expect(t, "76", int(outputs.Docs[2].Scores[0]*1000))
	utils.Expect(t, "[0 18]", outputs.Docs[2].TokenSnippetPositions)
}

func TestReverseOrder(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineInitOptions{
		SegmenterDictionaries: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			ReverseOrder:    true,
			OutputOffset:    0,
			MaxOutputs:      10,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.LocationsIndex,
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchRequest{Text: "中国人口"})
	utils.Expect(t, "3", len(outputs.Docs))

	utils.Expect(t, "0", outputs.Docs[0].DocId)
	utils.Expect(t, "4", outputs.Docs[1].DocId)
	utils.Expect(t, "1", outputs.Docs[2].DocId)
}

func TestOffsetAndMaxOutputs(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineInitOptions{
		SegmenterDictionaries: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			ReverseOrder:    true,
			OutputOffset:    1,
			MaxOutputs:      3,
			ScoringCriteria: &RankByTokenProximity{},
		},
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.LocationsIndex,
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchRequest{Text: "中国人口"})
	utils.Expect(t, "2", len(outputs.Docs))

	utils.Expect(t, "4", outputs.Docs[0].DocId)
	utils.Expect(t, "1", outputs.Docs[1].DocId)
}

type TestScoringCriteria struct {
}

func (criteria TestScoringCriteria) Score(
	doc types.IndexedDocument, fields interface{}) []float32 {
	if reflect.TypeOf(fields) != reflect.TypeOf(ScoringFields{}) {
		return []float32{}
	}
	fs := fields.(ScoringFields)
	return []float32{float32(doc.TokenProximity)*fs.a + fs.b*fs.c}
}

func TestSearchWithCriteria(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineInitOptions{
		SegmenterDictionaries: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			ScoringCriteria: TestScoringCriteria{},
		},
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.LocationsIndex,
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchRequest{Text: "中国人口"})
	utils.Expect(t, "2", len(outputs.Docs))

	utils.Expect(t, "0", outputs.Docs[0].DocId)
	utils.Expect(t, "18000", int(outputs.Docs[0].Scores[0]*1000))

	utils.Expect(t, "4", outputs.Docs[1].DocId)
	utils.Expect(t, "9000", int(outputs.Docs[1].Scores[0]*1000))
}

func TestCompactIndex(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineInitOptions{
		SegmenterDictionaries: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			ScoringCriteria: TestScoringCriteria{},
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchRequest{Text: "中国人口"})
	utils.Expect(t, "2", len(outputs.Docs))

	utils.Expect(t, "4", outputs.Docs[0].DocId)
	utils.Expect(t, "9000", int(outputs.Docs[0].Scores[0]*1000))

	utils.Expect(t, "0", outputs.Docs[1].DocId)
	utils.Expect(t, "6000", int(outputs.Docs[1].Scores[0]*1000))
}

type BM25ScoringCriteria struct {
}

func (criteria BM25ScoringCriteria) Score(
	doc types.IndexedDocument, fields interface{}) []float32 {
	if reflect.TypeOf(fields) != reflect.TypeOf(ScoringFields{}) {
		return []float32{}
	}
	return []float32{doc.BM25}
}

func TestFrequenciesIndex(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineInitOptions{
		SegmenterDictionaries: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			ScoringCriteria: BM25ScoringCriteria{},
		},
		IndexerInitOptions: &types.IndexerInitOptions{
			IndexType: types.FrequenciesIndex,
		},
	})

	AddDocs(&engine)

	outputs := engine.Search(types.SearchRequest{Text: "中国人口"})
	utils.Expect(t, "2", len(outputs.Docs))

	utils.Expect(t, "4", outputs.Docs[0].DocId)
	utils.Expect(t, "2311", int(outputs.Docs[0].Scores[0]*1000))

	utils.Expect(t, "0", outputs.Docs[1].DocId)
	utils.Expect(t, "2211", int(outputs.Docs[1].Scores[0]*1000))
}

func TestRemoveDocument(t *testing.T) {
	var engine Engine
	engine.Init(types.EngineInitOptions{
		SegmenterDictionaries: "../testdata/test_dict.txt",
		DefaultRankOptions: &types.RankOptions{
			ScoringCriteria: TestScoringCriteria{},
		},
	})

	AddDocs(&engine)
	engine.RemoveDocument(4)

	outputs := engine.Search(types.SearchRequest{Text: "中国人口"})
	utils.Expect(t, "1", len(outputs.Docs))

	utils.Expect(t, "0", outputs.Docs[0].DocId)
	utils.Expect(t, "6000", int(outputs.Docs[0].Scores[0]*1000))
}
