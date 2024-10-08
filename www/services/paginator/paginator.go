package paginator

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"math"
	"web-forum/internal"
	"web-forum/system"
	"web-forum/system/db"
)

var ctx = context.Background()

func Query(preQuery internal.PaginatorPreQuery) (tx pgx.Tx, rows pgx.Rows, paginatorList internal.Paginator, err error) {
	const errorFunction = "paginator.Query"

	tx, err = db.Postgres.Begin(ctx)

	if err != nil {
		return nil, nil, paginatorList, err
	}

	var topicsCount float64
	var tableName = preQuery.TableName
	var outputColumns = preQuery.OutputColumns
	var page = preQuery.Page
	var columnName = preQuery.WhereColumn
	var id = preQuery.WhereValue

	var orderStr string
	var orderDesc = preQuery.OrderReverse

	if orderDesc {
		orderStr = "desc"
	}

	if preQuery.QueryCount.PreparedValue != 0 {
		topicsCount = float64(preQuery.QueryCount.PreparedValue)
	} else {
		queryRow := tx.QueryRow(ctx, preQuery.QueryCount.Query)
		countMessagesErr := queryRow.Scan(&topicsCount)

		if countMessagesErr != nil {
			system.ErrLog(errorFunction, countMessagesErr)
			topicsCount = 1
		}
	}

	pagesCount := math.Ceil(topicsCount / internal.MaxPaginatorMessages)

	if float64(page) > pagesCount || float64(page) < 0 {
		page = 1
	}

	var whereStr string

	if preQuery.WhereOperator == "" {
		preQuery.WhereOperator = "="
	}

	if id != nil {
		whereStr = fmt.Sprintf("where %s %s $1", columnName, preQuery.WhereOperator)
	}

	fmtQuery := fmt.Sprintf(`select %s
	from %s
	where id in (
		select id from (
			select id, row_number() over(order by id)
			from %s
			%s
			offset %d
			limit %d
		)
		order by id %s
	)
	order by id %s;`, outputColumns, tableName, tableName, whereStr, (page-1)*internal.MaxPaginatorMessages, internal.MaxPaginatorMessages, orderStr, orderStr)

	if id != nil {
		rows, err = tx.Query(ctx, fmtQuery, id)
	} else {
		rows, err = tx.Query(ctx, fmtQuery)
	}

	if err != nil {
		return nil, nil, paginatorList, err
	}

	paginatorList.CurrentPage = page
	paginatorList.AllPages = int(pagesCount)

	Construct(&paginatorList)

	return tx, rows, paginatorList, nil
}

func Construct(paginatorList *internal.Paginator) {
	currentPageInt := (*paginatorList).CurrentPage
	ourPages := (*paginatorList).AllPages
	howMuchPagesWillBeVisible := internal.HowMuchPagesWillBeVisibleInPaginator
	dividedBy2 := float64(howMuchPagesWillBeVisible) / 2
	floorDivided := int(math.Floor(dividedBy2))
	ceilDivided := int(math.Ceil(dividedBy2))

	if ourPages < internal.HowMuchPagesWillBeVisibleInPaginator {
		howMuchPagesWillBeVisible = ourPages
	}

	if currentPageInt > ourPages {
		currentPageInt = ourPages
	}

	currentPageInt = currentPageInt - 1 // Массив с нуля начинается.
	limitMin, limitMax := currentPageInt-floorDivided, currentPageInt+floorDivided

	if limitMin < 0 {
		limitMin = 0
	}

	if limitMax > ourPages-1 {
		limitMax = ourPages - 1
	}

	if currentPageInt < ceilDivided {
		limitMax = howMuchPagesWillBeVisible - 1
	} else if currentPageInt >= ourPages-ceilDivided {
		limitMin = ourPages - howMuchPagesWillBeVisible
	}

	paginatorPages := make([]int, limitMax-limitMin+1)
	paginatorKey := 0

	for showedPage := limitMin; showedPage <= limitMax; showedPage++ {
		paginatorPages[paginatorKey] = showedPage + 1
		paginatorKey += 1
	}

	(*paginatorList).PagesArray = paginatorPages
	currentPageInt += 1

	if currentPageInt > 1 {
		(*paginatorList).Left.Activated = true
		(*paginatorList).Left.WhichPage = currentPageInt - 1
	}

	if currentPageInt < ourPages {
		(*paginatorList).Right.Activated = true
		(*paginatorList).Right.WhichPage = currentPageInt + 1
	}
}
