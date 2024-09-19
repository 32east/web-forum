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

func Query(tableName string, outputColumns string, columnName string, id int, page int, queryCount string) (tx pgx.Tx, rows pgx.Rows, paginatorMessages internal.Paginator, err error) {
	const errorFunction = "paginator.Query"

	tx, err = db.Postgres.Begin(ctx)

	if err != nil {
		return nil, nil, paginatorMessages, err
	}

	var topicsCount float64

	queryRow := tx.QueryRow(ctx, queryCount)
	countMessagesErr := queryRow.Scan(&topicsCount)

	if countMessagesErr != nil {
		system.ErrLog(errorFunction, countMessagesErr)
		topicsCount = 1
	}

	pagesCount := math.Ceil(topicsCount / internal.MaxPaginatorMessages)

	if float64(page) > pagesCount || float64(page) < 0 {
		page = 1
	}

	var whereStr string

	if id != -1 {
		whereStr = fmt.Sprintf("where %s = $1", columnName)
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
		order by id
	)
	order by id;`, outputColumns, tableName, tableName, whereStr, (page-1)*internal.MaxPaginatorMessages, internal.MaxPaginatorMessages)

	fmt.Println(fmtQuery)
	if id == -1 {
		rows, err = tx.Query(ctx, fmtQuery)
	} else {
		rows, err = tx.Query(ctx, fmtQuery, id)
	}

	if err != nil {
		return nil, nil, paginatorMessages, err
	}

	paginatorMessages.CurrentPage = page
	paginatorMessages.AllPages = int(pagesCount)

	return tx, rows, paginatorMessages, nil
}

func Construct(paginatorList internal.Paginator) internal.PaginatorConstructed {
	currentPageInt := paginatorList.CurrentPage
	ourPages := paginatorList.AllPages
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

	finalPaginator := internal.PaginatorConstructed{PagesArray: paginatorPages}
	currentPageInt += 1

	if currentPageInt > 1 {
		finalPaginator.Left.Activated = true
		finalPaginator.Left.WhichPage = currentPageInt - 1
	}

	if currentPageInt < ourPages {
		finalPaginator.Right.Activated = true
		finalPaginator.Right.WhichPage = currentPageInt + 1
	}

	return finalPaginator
}
