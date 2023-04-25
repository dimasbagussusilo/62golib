package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetFilterByQuery(query *gorm.DB, transformer map[string]any, ctx *gin.Context) map[string]any {
	filter := map[string]any{}
	queries := ctx.Request.URL.Query()

	if transformer["filterable"] != nil {
		filterable := transformer["filterable"].(map[string]any)

		for name, values := range queries {
			name = strings.Replace(name, "[]", "", -1)

			if _, ok := filterable[name]; ok {
				if values[0] != "" {
					query.Where(name+" IN ?", values)
				} else {
					query.Where(name + " IS NULL")
				}

				filter[name] = values
			}
		}
	}

	delete(transformer, "filterable")

	return filter
}

func SetPagination(query *gorm.DB, ctx *gin.Context) map[string]any {
	if page, _ := strconv.Atoi(ctx.Query("page")); page != 0 {
		var total int64

		if err := DB.Table(query.Statement.Table).Count(&total).Error; err != nil {
			fmt.Println(err)
		}

		per_page, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "30"))
		offset := (page - 1) * per_page
		query.Limit(per_page).Offset(offset)

		return map[string]any{
			"total":        total,
			"per_page":     per_page,
			"current_page": page,
			"last_page":    int(math.Ceil(float64(total) / float64(per_page))),
		}
	}

	return map[string]any{}
}

func SetBelongsTo(query *gorm.DB, transformer map[string]any, columns *[]string) {
	if transformer["belongs_to"] != nil {
		for _, v := range transformer["belongs_to"].(map[string]any) {
			v := v.(map[string]any)
			table := v["table"].(string)
			query.Joins("left join " + table + " on " + query.Statement.Table + "." + v["fk"].(string) + " = " + table + ".id")
			query.Select("products.height").Select("users.id")

			for _, val := range v["columns"].([]any) {
				*columns = append(*columns, table+"."+val.(string)+" as "+table+"_"+val.(string))
			}

		}
	}
}

func AttachHasMany(transformer map[string]any) {
	if transformer["has_many"] != nil {
		for i, v := range transformer["has_many"].(map[string]any) {
			v := v.(map[string]any)
			values := []map[string]any{}
			colums := convertAnyToString(v["columns"].([]any))
			fk := v["fk"].(string)

			if err := DB.Table(v["table"].(string)).Select(colums).Where(fk+" = ?", transformer["id"]).Find(&values).Error; err != nil {
				fmt.Println(err)
			}

			transformer[i] = values
		}
	}

	delete(transformer, "has_many")
}

func MultiAttachHasMany(results []map[string]any) {
	ids := []string{}

	for _, result := range results {
		ids = append(ids, strconv.Itoa(int(result["id"].(int32))))
	}

	transformer := results[0]

	if transformer["has_many"] != nil {
		for i, v := range transformer["has_many"].(map[string]any) {
			v := v.(map[string]any)
			values := []map[string]any{}
			fk := v["fk"].(string)
			colums := convertAnyToString(v["columns"].([]any))
			colums = append(colums, fk)

			if err := DB.Table(v["table"].(string)).Select(colums).Where(fk+" in ?", ids).Find(&values).Error; err != nil {
				fmt.Println(err)
			}

			for _, result := range results {
				result[i] = filterSliceByMapIndex(values, fk, result["id"])
			}
		}
	}

	delete(transformer, "has_many")
}

func AttachBelongsTo(transformer, value map[string]any) {
	if transformer["belongs_to"] != nil {
		for i, v := range transformer["belongs_to"].(map[string]any) {
			v := v.(map[string]any)
			values := map[string]any{}

			for _, val := range v["columns"].([]any) {
				values[val.(string)] = value[v["table"].(string)+"_"+val.(string)]
				//delete(transformer, v["fk"].(string))
			}

			transformer[i] = values
		}
	}

	delete(transformer, "belongs_to")
}
