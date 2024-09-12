package paginator

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
	"math"
	"web-forum/internal"
	"web-forum/system/db"
)

var ctx = context.Background()

func Query(tableName string, columnName string, id int, page int) (tx pgx.Tx, rows pgx.Rows, paginatorMessages internal.Paginator, err error) {
	const errorFunction = "paginator.Query"

	tx, err = db.Postgres.Begin(ctx)

	if err != nil {
		return nil, nil, paginatorMessages, err
	}

	var topicsCount float64
	fmtQuery := fmt.Sprintf("select count(*) from %s where %s = $1", tableName, columnName)
	queryRow := tx.QueryRow(ctx, fmtQuery, id)
	countMessagesErr := queryRow.Scan(&topicsCount)
	pagesCount := math.Ceil(topicsCount / internal.MaxPaginatorMessages)

	if countMessagesErr != nil {
		log.Fatal(fmt.Errorf("%s: %w", errorFunction, countMessagesErr))
	}

	fmtQuery = fmt.Sprintf("select * from %s where %s=$1 LIMIT %d OFFSET %d;", tableName, columnName, internal.MaxPaginatorMessages, (page-1)*internal.MaxPaginatorMessages)
	rows, err = tx.Query(ctx, fmtQuery, id)

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
