package paginator

import (
	"math"
	"web-forum/internal"
)

func PaginatorConstruct(paginatorList internal.Paginator) internal.PaginatorConstructed {
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
