package generator

import (
	response_utiliy "finance/utility/response"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Query struct {
	Limit  int
	Offset int
}

// Return Page Offset
func GenerateOffset(req *response_utiliy.StandarGetRequest) (Query, error) {
	var q Query
	q.Limit = 10
	q.Offset = 0

	if req.Page == "" {
		req.Page = "1"
	}

	if req.Record == "" {
		req.Record = "10"
	}

	iPage, err := strconv.Atoi(req.Page)
	if err != nil {
		return q, err
	}

	iRecord, err := strconv.Atoi(req.Record)
	if err != nil {
		return q, err
	}

	var dataOffset int = (iPage - 1) * iRecord
	q.Limit = iRecord
	q.Offset = dataOffset

	return q, nil
}

func GenerateSort(req *response_utiliy.StandarGetRequest) string {
	if req.Sort == "desc" {
		return "desc"
	} else {
		return "asc"
	}
}

func GenerateFilter(c *fiber.Ctx) *response_utiliy.StandarGetRequest {
	var filterData response_utiliy.StandarGetRequest
	filterData.Page = c.Query("page")
	filterData.OrderBy = c.Query("order-by")
	filterData.Search = c.Query("search")
	filterData.Record = c.Query("record")
	filterData.Sort = c.Query("sort")
	filterData.StartDate = c.Query("start-date")
	filterData.EndDate = c.Query("end-date")
	return &filterData
}
