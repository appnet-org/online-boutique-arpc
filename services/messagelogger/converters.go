package messagelogger

import (
	"fmt"

	pb "github.com/appnetorg/online-boutique-arpc/proto"
	pbcapnp "github.com/appnetorg/online-boutique-arpc/proto/capnp"
	fb "github.com/appnetorg/online-boutique-arpc/proto/proto"
	flatbuffers "github.com/google/flatbuffers/go"
	capnp "capnproto.org/go/capnp/v3"
)

// FlatBuffers converters

func ProtoToFB_Empty(pb *pb.Empty) ([]byte, error) {
	builder := flatbuffers.NewBuilder(64)
	fb.EmptyStart(builder)
	obj := fb.EmptyEnd(builder)
	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_EmptyUser(pbMsg *pb.EmptyUser) ([]byte, error) {
	builder := flatbuffers.NewBuilder(128)

	userID := builder.CreateString(pbMsg.GetUserId())

	fb.EmptyUserStart(builder)
	fb.EmptyUserAddUserId(builder, userID)
	obj := fb.EmptyUserEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_Money(pbMsg *pb.Money) ([]byte, error) {
	builder := flatbuffers.NewBuilder(128)

	currencyCode := builder.CreateString(pbMsg.GetCurrencyCode())

	fb.MoneyStart(builder)
	fb.MoneyAddCurrencyCode(builder, currencyCode)
	fb.MoneyAddUnits(builder, pbMsg.GetUnits())
	fb.MoneyAddNanos(builder, pbMsg.GetNanos())
	obj := fb.MoneyEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_Address(pbMsg *pb.Address) ([]byte, error) {
	builder := flatbuffers.NewBuilder(256)

	streetAddress := builder.CreateString(pbMsg.GetStreetAddress())
	city := builder.CreateString(pbMsg.GetCity())
	state := builder.CreateString(pbMsg.GetState())
	country := builder.CreateString(pbMsg.GetCountry())

	fb.AddressStart(builder)
	fb.AddressAddStreetAddress(builder, streetAddress)
	fb.AddressAddCity(builder, city)
	fb.AddressAddState(builder, state)
	fb.AddressAddCountry(builder, country)
	fb.AddressAddZipCode(builder, pbMsg.GetZipCode())
	obj := fb.AddressEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_CartItem(pbMsg *pb.CartItem) ([]byte, error) {
	builder := flatbuffers.NewBuilder(128)

	productID := builder.CreateString(pbMsg.GetProductId())

	fb.CartItemStart(builder)
	fb.CartItemAddProductId(builder, productID)
	fb.CartItemAddQuantity(builder, pbMsg.GetQuantity())
	obj := fb.CartItemEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_CreditCardInfo(pbMsg *pb.CreditCardInfo) ([]byte, error) {
	builder := flatbuffers.NewBuilder(128)

	cardNumber := builder.CreateString(pbMsg.GetCreditCardNumber())

	fb.CreditCardInfoStart(builder)
	fb.CreditCardInfoAddCreditCardNumber(builder, cardNumber)
	fb.CreditCardInfoAddCreditCardCvv(builder, pbMsg.GetCreditCardCvv())
	fb.CreditCardInfoAddCreditCardExpirationYear(builder, pbMsg.GetCreditCardExpirationYear())
	fb.CreditCardInfoAddCreditCardExpirationMonth(builder, pbMsg.GetCreditCardExpirationMonth())
	obj := fb.CreditCardInfoEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_Ad(pbMsg *pb.Ad) ([]byte, error) {
	builder := flatbuffers.NewBuilder(256)

	redirectURL := builder.CreateString(pbMsg.GetRedirectUrl())
	text := builder.CreateString(pbMsg.GetText())

	fb.AdStart(builder)
	fb.AdAddRedirectUrl(builder, redirectURL)
	fb.AdAddText(builder, text)
	obj := fb.AdEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_AddItemRequest(pbMsg *pb.AddItemRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(256)

	userID := builder.CreateString(pbMsg.GetUserId())

	// Build CartItem
	item := pbMsg.GetItem()
	productID := builder.CreateString(item.GetProductId())
	fb.CartItemStart(builder)
	fb.CartItemAddProductId(builder, productID)
	fb.CartItemAddQuantity(builder, item.GetQuantity())
	cartItem := fb.CartItemEnd(builder)

	fb.AddItemRequestStart(builder)
	fb.AddItemRequestAddUserId(builder, userID)
	fb.AddItemRequestAddItem(builder, cartItem)
	obj := fb.AddItemRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_GetCartRequest(pbMsg *pb.GetCartRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(128)

	userID := builder.CreateString(pbMsg.GetUserId())

	fb.GetCartRequestStart(builder)
	fb.GetCartRequestAddUserId(builder, userID)
	obj := fb.GetCartRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_EmptyCartRequest(pbMsg *pb.EmptyCartRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(128)

	userID := builder.CreateString(pbMsg.GetUserId())

	fb.EmptyCartRequestStart(builder)
	fb.EmptyCartRequestAddUserId(builder, userID)
	obj := fb.EmptyCartRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_Cart(pbMsg *pb.Cart) ([]byte, error) {
	builder := flatbuffers.NewBuilder(1024)

	userID := builder.CreateString(pbMsg.GetUserId())

	// Build CartItem list
	items := pbMsg.GetItems()
	itemOffsets := make([]flatbuffers.UOffsetT, len(items))
	for i, item := range items {
		productID := builder.CreateString(item.GetProductId())
		fb.CartItemStart(builder)
		fb.CartItemAddProductId(builder, productID)
		fb.CartItemAddQuantity(builder, item.GetQuantity())
		itemOffsets[i] = fb.CartItemEnd(builder)
	}
	fb.CartStartItemsVector(builder, len(itemOffsets))
	for i := len(itemOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(itemOffsets[i])
	}
	itemsVector := builder.EndVector(len(itemOffsets))

	fb.CartStart(builder)
	fb.CartAddUserId(builder, userID)
	fb.CartAddItems(builder, itemsVector)
	obj := fb.CartEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_ListRecommendationsRequest(pbMsg *pb.ListRecommendationsRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(512)

	userID := builder.CreateString(pbMsg.GetUserId())

	// Build product IDs list
	productIDs := pbMsg.GetProductIds()
	productIDOffsets := make([]flatbuffers.UOffsetT, len(productIDs))
	for i, pid := range productIDs {
		productIDOffsets[i] = builder.CreateString(pid)
	}
	fb.ListRecommendationsRequestStartProductIdsVector(builder, len(productIDOffsets))
	for i := len(productIDOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(productIDOffsets[i])
	}
	productIDsVector := builder.EndVector(len(productIDOffsets))

	fb.ListRecommendationsRequestStart(builder)
	fb.ListRecommendationsRequestAddUserId(builder, userID)
	fb.ListRecommendationsRequestAddProductIds(builder, productIDsVector)
	obj := fb.ListRecommendationsRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_ListRecommendationsResponse(pbMsg *pb.ListRecommendationsResponse) ([]byte, error) {
	builder := flatbuffers.NewBuilder(512)

	// Build product IDs list
	productIDs := pbMsg.GetProductIds()
	productIDOffsets := make([]flatbuffers.UOffsetT, len(productIDs))
	for i, pid := range productIDs {
		productIDOffsets[i] = builder.CreateString(pid)
	}
	fb.ListRecommendationsResponseStartProductIdsVector(builder, len(productIDOffsets))
	for i := len(productIDOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(productIDOffsets[i])
	}
	productIDsVector := builder.EndVector(len(productIDOffsets))

	fb.ListRecommendationsResponseStart(builder)
	fb.ListRecommendationsResponseAddProductIds(builder, productIDsVector)
	obj := fb.ListRecommendationsResponseEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_Product(pbMsg *pb.Product) ([]byte, error) {
	builder := flatbuffers.NewBuilder(1024)

	id := builder.CreateString(pbMsg.GetId())
	name := builder.CreateString(pbMsg.GetName())
	description := builder.CreateString(pbMsg.GetDescription())
	picture := builder.CreateString(pbMsg.GetPicture())

	// Build Money
	priceUsd := pbMsg.GetPriceUsd()
	currencyCode := builder.CreateString(priceUsd.GetCurrencyCode())
	fb.MoneyStart(builder)
	fb.MoneyAddCurrencyCode(builder, currencyCode)
	fb.MoneyAddUnits(builder, priceUsd.GetUnits())
	fb.MoneyAddNanos(builder, priceUsd.GetNanos())
	moneyOffset := fb.MoneyEnd(builder)

	// Build categories list
	categories := pbMsg.GetCategories()
	categoryOffsets := make([]flatbuffers.UOffsetT, len(categories))
	for i, cat := range categories {
		categoryOffsets[i] = builder.CreateString(cat)
	}
	fb.ProductStartCategoriesVector(builder, len(categoryOffsets))
	for i := len(categoryOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(categoryOffsets[i])
	}
	categoriesVector := builder.EndVector(len(categoryOffsets))

	fb.ProductStart(builder)
	fb.ProductAddId(builder, id)
	fb.ProductAddName(builder, name)
	fb.ProductAddDescription(builder, description)
	fb.ProductAddPicture(builder, picture)
	fb.ProductAddPriceUsd(builder, moneyOffset)
	fb.ProductAddCategories(builder, categoriesVector)
	obj := fb.ProductEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_GetProductRequest(pbMsg *pb.GetProductRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(128)

	id := builder.CreateString(pbMsg.GetId())

	fb.GetProductRequestStart(builder)
	fb.GetProductRequestAddId(builder, id)
	obj := fb.GetProductRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_SearchProductsRequest(pbMsg *pb.SearchProductsRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(256)

	query := builder.CreateString(pbMsg.GetQuery())

	fb.SearchProductsRequestStart(builder)
	fb.SearchProductsRequestAddQuery(builder, query)
	obj := fb.SearchProductsRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_ListProductsResponse(pbMsg *pb.ListProductsResponse) ([]byte, error) {
	builder := flatbuffers.NewBuilder(4096)

	// Build product list
	products := pbMsg.GetProducts()
	productOffsets := make([]flatbuffers.UOffsetT, len(products))

	for i, prod := range products {
		id := builder.CreateString(prod.GetId())
		name := builder.CreateString(prod.GetName())
		description := builder.CreateString(prod.GetDescription())
		picture := builder.CreateString(prod.GetPicture())

		// Build Money
		priceUsd := prod.GetPriceUsd()
		currencyCode := builder.CreateString(priceUsd.GetCurrencyCode())
		fb.MoneyStart(builder)
		fb.MoneyAddCurrencyCode(builder, currencyCode)
		fb.MoneyAddUnits(builder, priceUsd.GetUnits())
		fb.MoneyAddNanos(builder, priceUsd.GetNanos())
		moneyOffset := fb.MoneyEnd(builder)

		// Build categories
		categories := prod.GetCategories()
		categoryOffsets := make([]flatbuffers.UOffsetT, len(categories))
		for j, cat := range categories {
			categoryOffsets[j] = builder.CreateString(cat)
		}
		fb.ProductStartCategoriesVector(builder, len(categoryOffsets))
		for j := len(categoryOffsets) - 1; j >= 0; j-- {
			builder.PrependUOffsetT(categoryOffsets[j])
		}
		categoriesVector := builder.EndVector(len(categoryOffsets))

		fb.ProductStart(builder)
		fb.ProductAddId(builder, id)
		fb.ProductAddName(builder, name)
		fb.ProductAddDescription(builder, description)
		fb.ProductAddPicture(builder, picture)
		fb.ProductAddPriceUsd(builder, moneyOffset)
		fb.ProductAddCategories(builder, categoriesVector)
		productOffsets[i] = fb.ProductEnd(builder)
	}

	fb.ListProductsResponseStartProductsVector(builder, len(productOffsets))
	for i := len(productOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(productOffsets[i])
	}
	productsVector := builder.EndVector(len(productOffsets))

	fb.ListProductsResponseStart(builder)
	fb.ListProductsResponseAddProducts(builder, productsVector)
	obj := fb.ListProductsResponseEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_SearchProductsResponse(pbMsg *pb.SearchProductsResponse) ([]byte, error) {
	builder := flatbuffers.NewBuilder(4096)

	// Build results list (same as products)
	results := pbMsg.GetResults()
	resultOffsets := make([]flatbuffers.UOffsetT, len(results))

	for i, prod := range results {
		id := builder.CreateString(prod.GetId())
		name := builder.CreateString(prod.GetName())
		description := builder.CreateString(prod.GetDescription())
		picture := builder.CreateString(prod.GetPicture())

		// Build Money
		priceUsd := prod.GetPriceUsd()
		currencyCode := builder.CreateString(priceUsd.GetCurrencyCode())
		fb.MoneyStart(builder)
		fb.MoneyAddCurrencyCode(builder, currencyCode)
		fb.MoneyAddUnits(builder, priceUsd.GetUnits())
		fb.MoneyAddNanos(builder, priceUsd.GetNanos())
		moneyOffset := fb.MoneyEnd(builder)

		// Build categories
		categories := prod.GetCategories()
		categoryOffsets := make([]flatbuffers.UOffsetT, len(categories))
		for j, cat := range categories {
			categoryOffsets[j] = builder.CreateString(cat)
		}
		fb.ProductStartCategoriesVector(builder, len(categoryOffsets))
		for j := len(categoryOffsets) - 1; j >= 0; j-- {
			builder.PrependUOffsetT(categoryOffsets[j])
		}
		categoriesVector := builder.EndVector(len(categoryOffsets))

		fb.ProductStart(builder)
		fb.ProductAddId(builder, id)
		fb.ProductAddName(builder, name)
		fb.ProductAddDescription(builder, description)
		fb.ProductAddPicture(builder, picture)
		fb.ProductAddPriceUsd(builder, moneyOffset)
		fb.ProductAddCategories(builder, categoriesVector)
		resultOffsets[i] = fb.ProductEnd(builder)
	}

	fb.SearchProductsResponseStartResultsVector(builder, len(resultOffsets))
	for i := len(resultOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(resultOffsets[i])
	}
	resultsVector := builder.EndVector(len(resultOffsets))

	fb.SearchProductsResponseStart(builder)
	fb.SearchProductsResponseAddResults(builder, resultsVector)
	obj := fb.SearchProductsResponseEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_GetQuoteRequest(pbMsg *pb.GetQuoteRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(1024)

	// Build Address
	addr := pbMsg.GetAddress()
	streetAddress := builder.CreateString(addr.GetStreetAddress())
	city := builder.CreateString(addr.GetCity())
	state := builder.CreateString(addr.GetState())
	country := builder.CreateString(addr.GetCountry())
	fb.AddressStart(builder)
	fb.AddressAddStreetAddress(builder, streetAddress)
	fb.AddressAddCity(builder, city)
	fb.AddressAddState(builder, state)
	fb.AddressAddCountry(builder, country)
	fb.AddressAddZipCode(builder, addr.GetZipCode())
	addressOffset := fb.AddressEnd(builder)

	// Build CartItem list
	items := pbMsg.GetItems()
	itemOffsets := make([]flatbuffers.UOffsetT, len(items))
	for i, item := range items {
		productID := builder.CreateString(item.GetProductId())
		fb.CartItemStart(builder)
		fb.CartItemAddProductId(builder, productID)
		fb.CartItemAddQuantity(builder, item.GetQuantity())
		itemOffsets[i] = fb.CartItemEnd(builder)
	}
	fb.GetQuoteRequestStartItemsVector(builder, len(itemOffsets))
	for i := len(itemOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(itemOffsets[i])
	}
	itemsVector := builder.EndVector(len(itemOffsets))

	fb.GetQuoteRequestStart(builder)
	fb.GetQuoteRequestAddAddress(builder, addressOffset)
	fb.GetQuoteRequestAddItems(builder, itemsVector)
	obj := fb.GetQuoteRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_GetQuoteResponse(pbMsg *pb.GetQuoteResponse) ([]byte, error) {
	builder := flatbuffers.NewBuilder(256)

	// Build Money
	costUsd := pbMsg.GetCostUsd()
	currencyCode := builder.CreateString(costUsd.GetCurrencyCode())
	fb.MoneyStart(builder)
	fb.MoneyAddCurrencyCode(builder, currencyCode)
	fb.MoneyAddUnits(builder, costUsd.GetUnits())
	fb.MoneyAddNanos(builder, costUsd.GetNanos())
	moneyOffset := fb.MoneyEnd(builder)

	fb.GetQuoteResponseStart(builder)
	fb.GetQuoteResponseAddCostUsd(builder, moneyOffset)
	obj := fb.GetQuoteResponseEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_ShipOrderRequest(pbMsg *pb.ShipOrderRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(1024)

	// Build Address
	addr := pbMsg.GetAddress()
	streetAddress := builder.CreateString(addr.GetStreetAddress())
	city := builder.CreateString(addr.GetCity())
	state := builder.CreateString(addr.GetState())
	country := builder.CreateString(addr.GetCountry())
	fb.AddressStart(builder)
	fb.AddressAddStreetAddress(builder, streetAddress)
	fb.AddressAddCity(builder, city)
	fb.AddressAddState(builder, state)
	fb.AddressAddCountry(builder, country)
	fb.AddressAddZipCode(builder, addr.GetZipCode())
	addressOffset := fb.AddressEnd(builder)

	// Build CartItem list
	items := pbMsg.GetItems()
	itemOffsets := make([]flatbuffers.UOffsetT, len(items))
	for i, item := range items {
		productID := builder.CreateString(item.GetProductId())
		fb.CartItemStart(builder)
		fb.CartItemAddProductId(builder, productID)
		fb.CartItemAddQuantity(builder, item.GetQuantity())
		itemOffsets[i] = fb.CartItemEnd(builder)
	}
	fb.ShipOrderRequestStartItemsVector(builder, len(itemOffsets))
	for i := len(itemOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(itemOffsets[i])
	}
	itemsVector := builder.EndVector(len(itemOffsets))

	fb.ShipOrderRequestStart(builder)
	fb.ShipOrderRequestAddAddress(builder, addressOffset)
	fb.ShipOrderRequestAddItems(builder, itemsVector)
	obj := fb.ShipOrderRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_ShipOrderResponse(pbMsg *pb.ShipOrderResponse) ([]byte, error) {
	builder := flatbuffers.NewBuilder(128)

	trackingID := builder.CreateString(pbMsg.GetTrackingId())

	fb.ShipOrderResponseStart(builder)
	fb.ShipOrderResponseAddTrackingId(builder, trackingID)
	obj := fb.ShipOrderResponseEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_GetSupportedCurrenciesResponse(pbMsg *pb.GetSupportedCurrenciesResponse) ([]byte, error) {
	builder := flatbuffers.NewBuilder(512)

	// Build currency codes list
	currencyCodes := pbMsg.GetCurrencyCodes()
	codeOffsets := make([]flatbuffers.UOffsetT, len(currencyCodes))
	for i, code := range currencyCodes {
		codeOffsets[i] = builder.CreateString(code)
	}
	fb.GetSupportedCurrenciesResponseStartCurrencyCodesVector(builder, len(codeOffsets))
	for i := len(codeOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(codeOffsets[i])
	}
	codesVector := builder.EndVector(len(codeOffsets))

	fb.GetSupportedCurrenciesResponseStart(builder)
	fb.GetSupportedCurrenciesResponseAddCurrencyCodes(builder, codesVector)
	obj := fb.GetSupportedCurrenciesResponseEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_CurrencyConversionRequest(pbMsg *pb.CurrencyConversionRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(256)

	// Build Money
	from := pbMsg.GetFrom()
	currencyCode := builder.CreateString(from.GetCurrencyCode())
	fb.MoneyStart(builder)
	fb.MoneyAddCurrencyCode(builder, currencyCode)
	fb.MoneyAddUnits(builder, from.GetUnits())
	fb.MoneyAddNanos(builder, from.GetNanos())
	moneyOffset := fb.MoneyEnd(builder)

	toCode := builder.CreateString(pbMsg.GetToCode())
	userID := builder.CreateString(pbMsg.GetUserId())

	fb.CurrencyConversionRequestStart(builder)
	fb.CurrencyConversionRequestAddFrom(builder, moneyOffset)
	fb.CurrencyConversionRequestAddToCode(builder, toCode)
	fb.CurrencyConversionRequestAddUserId(builder, userID)
	obj := fb.CurrencyConversionRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_ChargeRequest(pbMsg *pb.ChargeRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(256)

	// Build Money
	amount := pbMsg.GetAmount()
	currencyCode := builder.CreateString(amount.GetCurrencyCode())
	fb.MoneyStart(builder)
	fb.MoneyAddCurrencyCode(builder, currencyCode)
	fb.MoneyAddUnits(builder, amount.GetUnits())
	fb.MoneyAddNanos(builder, amount.GetNanos())
	moneyOffset := fb.MoneyEnd(builder)

	// Build CreditCardInfo
	cc := pbMsg.GetCreditCard()
	cardNumber := builder.CreateString(cc.GetCreditCardNumber())
	fb.CreditCardInfoStart(builder)
	fb.CreditCardInfoAddCreditCardNumber(builder, cardNumber)
	fb.CreditCardInfoAddCreditCardCvv(builder, cc.GetCreditCardCvv())
	fb.CreditCardInfoAddCreditCardExpirationYear(builder, cc.GetCreditCardExpirationYear())
	fb.CreditCardInfoAddCreditCardExpirationMonth(builder, cc.GetCreditCardExpirationMonth())
	ccOffset := fb.CreditCardInfoEnd(builder)

	fb.ChargeRequestStart(builder)
	fb.ChargeRequestAddAmount(builder, moneyOffset)
	fb.ChargeRequestAddCreditCard(builder, ccOffset)
	obj := fb.ChargeRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_ChargeResponse(pbMsg *pb.ChargeResponse) ([]byte, error) {
	builder := flatbuffers.NewBuilder(128)

	transactionID := builder.CreateString(pbMsg.GetTransactionId())

	fb.ChargeResponseStart(builder)
	fb.ChargeResponseAddTransactionId(builder, transactionID)
	obj := fb.ChargeResponseEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_OrderItem(pbMsg *pb.OrderItem) ([]byte, error) {
	builder := flatbuffers.NewBuilder(256)

	// Build CartItem
	item := pbMsg.GetItem()
	productID := builder.CreateString(item.GetProductId())
	fb.CartItemStart(builder)
	fb.CartItemAddProductId(builder, productID)
	fb.CartItemAddQuantity(builder, item.GetQuantity())
	cartItemOffset := fb.CartItemEnd(builder)

	// Build Money
	cost := pbMsg.GetCost()
	currencyCode := builder.CreateString(cost.GetCurrencyCode())
	fb.MoneyStart(builder)
	fb.MoneyAddCurrencyCode(builder, currencyCode)
	fb.MoneyAddUnits(builder, cost.GetUnits())
	fb.MoneyAddNanos(builder, cost.GetNanos())
	moneyOffset := fb.MoneyEnd(builder)

	fb.OrderItemStart(builder)
	fb.OrderItemAddItem(builder, cartItemOffset)
	fb.OrderItemAddCost(builder, moneyOffset)
	obj := fb.OrderItemEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_OrderResult(pbMsg *pb.OrderResult) ([]byte, error) {
	builder := flatbuffers.NewBuilder(2048)

	orderID := builder.CreateString(pbMsg.GetOrderId())
	shippingTrackingID := builder.CreateString(pbMsg.GetShippingTrackingId())

	// Build Money (shipping cost)
	shippingCost := pbMsg.GetShippingCost()
	currencyCode := builder.CreateString(shippingCost.GetCurrencyCode())
	fb.MoneyStart(builder)
	fb.MoneyAddCurrencyCode(builder, currencyCode)
	fb.MoneyAddUnits(builder, shippingCost.GetUnits())
	fb.MoneyAddNanos(builder, shippingCost.GetNanos())
	moneyOffset := fb.MoneyEnd(builder)

	// Build Address
	addr := pbMsg.GetShippingAddress()
	streetAddress := builder.CreateString(addr.GetStreetAddress())
	city := builder.CreateString(addr.GetCity())
	state := builder.CreateString(addr.GetState())
	country := builder.CreateString(addr.GetCountry())
	fb.AddressStart(builder)
	fb.AddressAddStreetAddress(builder, streetAddress)
	fb.AddressAddCity(builder, city)
	fb.AddressAddState(builder, state)
	fb.AddressAddCountry(builder, country)
	fb.AddressAddZipCode(builder, addr.GetZipCode())
	addressOffset := fb.AddressEnd(builder)

	// Build OrderItem list
	items := pbMsg.GetItems()
	itemOffsets := make([]flatbuffers.UOffsetT, len(items))
	for i, orderItem := range items {
		// Build CartItem
		cartItem := orderItem.GetItem()
		productID := builder.CreateString(cartItem.GetProductId())
		fb.CartItemStart(builder)
		fb.CartItemAddProductId(builder, productID)
		fb.CartItemAddQuantity(builder, cartItem.GetQuantity())
		cartItemOff := fb.CartItemEnd(builder)

		// Build Money
		cost := orderItem.GetCost()
		cc := builder.CreateString(cost.GetCurrencyCode())
		fb.MoneyStart(builder)
		fb.MoneyAddCurrencyCode(builder, cc)
		fb.MoneyAddUnits(builder, cost.GetUnits())
		fb.MoneyAddNanos(builder, cost.GetNanos())
		moneyOff := fb.MoneyEnd(builder)

		fb.OrderItemStart(builder)
		fb.OrderItemAddItem(builder, cartItemOff)
		fb.OrderItemAddCost(builder, moneyOff)
		itemOffsets[i] = fb.OrderItemEnd(builder)
	}
	fb.OrderResultStartItemsVector(builder, len(itemOffsets))
	for i := len(itemOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(itemOffsets[i])
	}
	itemsVector := builder.EndVector(len(itemOffsets))

	fb.OrderResultStart(builder)
	fb.OrderResultAddOrderId(builder, orderID)
	fb.OrderResultAddShippingTrackingId(builder, shippingTrackingID)
	fb.OrderResultAddShippingCost(builder, moneyOffset)
	fb.OrderResultAddShippingAddress(builder, addressOffset)
	fb.OrderResultAddItems(builder, itemsVector)
	obj := fb.OrderResultEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_SendOrderConfirmationRequest(pbMsg *pb.SendOrderConfirmationRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(2048)

	email := builder.CreateString(pbMsg.GetEmail())

	// Build OrderResult
	order := pbMsg.GetOrder()
	orderID := builder.CreateString(order.GetOrderId())
	shippingTrackingID := builder.CreateString(order.GetShippingTrackingId())

	// Build Money (shipping cost)
	shippingCost := order.GetShippingCost()
	currencyCode := builder.CreateString(shippingCost.GetCurrencyCode())
	fb.MoneyStart(builder)
	fb.MoneyAddCurrencyCode(builder, currencyCode)
	fb.MoneyAddUnits(builder, shippingCost.GetUnits())
	fb.MoneyAddNanos(builder, shippingCost.GetNanos())
	moneyOffset := fb.MoneyEnd(builder)

	// Build Address
	addr := order.GetShippingAddress()
	streetAddress := builder.CreateString(addr.GetStreetAddress())
	city := builder.CreateString(addr.GetCity())
	state := builder.CreateString(addr.GetState())
	country := builder.CreateString(addr.GetCountry())
	fb.AddressStart(builder)
	fb.AddressAddStreetAddress(builder, streetAddress)
	fb.AddressAddCity(builder, city)
	fb.AddressAddState(builder, state)
	fb.AddressAddCountry(builder, country)
	fb.AddressAddZipCode(builder, addr.GetZipCode())
	addressOffset := fb.AddressEnd(builder)

	// Build OrderItem list
	items := order.GetItems()
	itemOffsets := make([]flatbuffers.UOffsetT, len(items))
	for i, orderItem := range items {
		// Build CartItem
		cartItem := orderItem.GetItem()
		productID := builder.CreateString(cartItem.GetProductId())
		fb.CartItemStart(builder)
		fb.CartItemAddProductId(builder, productID)
		fb.CartItemAddQuantity(builder, cartItem.GetQuantity())
		cartItemOff := fb.CartItemEnd(builder)

		// Build Money
		cost := orderItem.GetCost()
		cc := builder.CreateString(cost.GetCurrencyCode())
		fb.MoneyStart(builder)
		fb.MoneyAddCurrencyCode(builder, cc)
		fb.MoneyAddUnits(builder, cost.GetUnits())
		fb.MoneyAddNanos(builder, cost.GetNanos())
		moneyOff := fb.MoneyEnd(builder)

		fb.OrderItemStart(builder)
		fb.OrderItemAddItem(builder, cartItemOff)
		fb.OrderItemAddCost(builder, moneyOff)
		itemOffsets[i] = fb.OrderItemEnd(builder)
	}
	fb.OrderResultStartItemsVector(builder, len(itemOffsets))
	for i := len(itemOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(itemOffsets[i])
	}
	itemsVector := builder.EndVector(len(itemOffsets))

	fb.OrderResultStart(builder)
	fb.OrderResultAddOrderId(builder, orderID)
	fb.OrderResultAddShippingTrackingId(builder, shippingTrackingID)
	fb.OrderResultAddShippingCost(builder, moneyOffset)
	fb.OrderResultAddShippingAddress(builder, addressOffset)
	fb.OrderResultAddItems(builder, itemsVector)
	orderOffset := fb.OrderResultEnd(builder)

	fb.SendOrderConfirmationRequestStart(builder)
	fb.SendOrderConfirmationRequestAddEmail(builder, email)
	fb.SendOrderConfirmationRequestAddOrder(builder, orderOffset)
	obj := fb.SendOrderConfirmationRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_PlaceOrderRequest(pbMsg *pb.PlaceOrderRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(1024)

	userID := builder.CreateString(pbMsg.GetUserId())
	userCurrency := builder.CreateString(pbMsg.GetUserCurrency())
	email := builder.CreateString(pbMsg.GetEmail())

	// Build Address
	addr := pbMsg.GetAddress()
	streetAddress := builder.CreateString(addr.GetStreetAddress())
	city := builder.CreateString(addr.GetCity())
	state := builder.CreateString(addr.GetState())
	country := builder.CreateString(addr.GetCountry())
	fb.AddressStart(builder)
	fb.AddressAddStreetAddress(builder, streetAddress)
	fb.AddressAddCity(builder, city)
	fb.AddressAddState(builder, state)
	fb.AddressAddCountry(builder, country)
	fb.AddressAddZipCode(builder, addr.GetZipCode())
	addressOffset := fb.AddressEnd(builder)

	// Build CreditCardInfo
	cc := pbMsg.GetCreditCard()
	cardNumber := builder.CreateString(cc.GetCreditCardNumber())
	fb.CreditCardInfoStart(builder)
	fb.CreditCardInfoAddCreditCardNumber(builder, cardNumber)
	fb.CreditCardInfoAddCreditCardCvv(builder, cc.GetCreditCardCvv())
	fb.CreditCardInfoAddCreditCardExpirationYear(builder, cc.GetCreditCardExpirationYear())
	fb.CreditCardInfoAddCreditCardExpirationMonth(builder, cc.GetCreditCardExpirationMonth())
	ccOffset := fb.CreditCardInfoEnd(builder)

	fb.PlaceOrderRequestStart(builder)
	fb.PlaceOrderRequestAddUserId(builder, userID)
	fb.PlaceOrderRequestAddUserCurrency(builder, userCurrency)
	fb.PlaceOrderRequestAddAddress(builder, addressOffset)
	fb.PlaceOrderRequestAddEmail(builder, email)
	fb.PlaceOrderRequestAddCreditCard(builder, ccOffset)
	obj := fb.PlaceOrderRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_PlaceOrderResponse(pbMsg *pb.PlaceOrderResponse) ([]byte, error) {
	builder := flatbuffers.NewBuilder(2048)

	// Build OrderResult
	order := pbMsg.GetOrder()
	orderID := builder.CreateString(order.GetOrderId())
	shippingTrackingID := builder.CreateString(order.GetShippingTrackingId())

	// Build Money (shipping cost)
	shippingCost := order.GetShippingCost()
	currencyCode := builder.CreateString(shippingCost.GetCurrencyCode())
	fb.MoneyStart(builder)
	fb.MoneyAddCurrencyCode(builder, currencyCode)
	fb.MoneyAddUnits(builder, shippingCost.GetUnits())
	fb.MoneyAddNanos(builder, shippingCost.GetNanos())
	moneyOffset := fb.MoneyEnd(builder)

	// Build Address
	addr := order.GetShippingAddress()
	streetAddress := builder.CreateString(addr.GetStreetAddress())
	city := builder.CreateString(addr.GetCity())
	state := builder.CreateString(addr.GetState())
	country := builder.CreateString(addr.GetCountry())
	fb.AddressStart(builder)
	fb.AddressAddStreetAddress(builder, streetAddress)
	fb.AddressAddCity(builder, city)
	fb.AddressAddState(builder, state)
	fb.AddressAddCountry(builder, country)
	fb.AddressAddZipCode(builder, addr.GetZipCode())
	addressOffset := fb.AddressEnd(builder)

	// Build OrderItem list
	items := order.GetItems()
	itemOffsets := make([]flatbuffers.UOffsetT, len(items))
	for i, orderItem := range items {
		// Build CartItem
		cartItem := orderItem.GetItem()
		productID := builder.CreateString(cartItem.GetProductId())
		fb.CartItemStart(builder)
		fb.CartItemAddProductId(builder, productID)
		fb.CartItemAddQuantity(builder, cartItem.GetQuantity())
		cartItemOff := fb.CartItemEnd(builder)

		// Build Money
		cost := orderItem.GetCost()
		cc := builder.CreateString(cost.GetCurrencyCode())
		fb.MoneyStart(builder)
		fb.MoneyAddCurrencyCode(builder, cc)
		fb.MoneyAddUnits(builder, cost.GetUnits())
		fb.MoneyAddNanos(builder, cost.GetNanos())
		moneyOff := fb.MoneyEnd(builder)

		fb.OrderItemStart(builder)
		fb.OrderItemAddItem(builder, cartItemOff)
		fb.OrderItemAddCost(builder, moneyOff)
		itemOffsets[i] = fb.OrderItemEnd(builder)
	}
	fb.OrderResultStartItemsVector(builder, len(itemOffsets))
	for i := len(itemOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(itemOffsets[i])
	}
	itemsVector := builder.EndVector(len(itemOffsets))

	fb.OrderResultStart(builder)
	fb.OrderResultAddOrderId(builder, orderID)
	fb.OrderResultAddShippingTrackingId(builder, shippingTrackingID)
	fb.OrderResultAddShippingCost(builder, moneyOffset)
	fb.OrderResultAddShippingAddress(builder, addressOffset)
	fb.OrderResultAddItems(builder, itemsVector)
	orderOffset := fb.OrderResultEnd(builder)

	fb.PlaceOrderResponseStart(builder)
	fb.PlaceOrderResponseAddOrder(builder, orderOffset)
	obj := fb.PlaceOrderResponseEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_AdRequest(pbMsg *pb.AdRequest) ([]byte, error) {
	builder := flatbuffers.NewBuilder(512)

	userID := builder.CreateString(pbMsg.GetUserId())

	// Build context keys list
	contextKeys := pbMsg.GetContextKeys()
	keyOffsets := make([]flatbuffers.UOffsetT, len(contextKeys))
	for i, key := range contextKeys {
		keyOffsets[i] = builder.CreateString(key)
	}
	fb.AdRequestStartContextKeysVector(builder, len(keyOffsets))
	for i := len(keyOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(keyOffsets[i])
	}
	keysVector := builder.EndVector(len(keyOffsets))

	fb.AdRequestStart(builder)
	fb.AdRequestAddUserId(builder, userID)
	fb.AdRequestAddContextKeys(builder, keysVector)
	obj := fb.AdRequestEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

func ProtoToFB_AdResponse(pbMsg *pb.AdResponse) ([]byte, error) {
	builder := flatbuffers.NewBuilder(1024)

	// Build ads list
	ads := pbMsg.GetAds()
	adOffsets := make([]flatbuffers.UOffsetT, len(ads))
	for i, ad := range ads {
		redirectURL := builder.CreateString(ad.GetRedirectUrl())
		text := builder.CreateString(ad.GetText())

		fb.AdStart(builder)
		fb.AdAddRedirectUrl(builder, redirectURL)
		fb.AdAddText(builder, text)
		adOffsets[i] = fb.AdEnd(builder)
	}
	fb.AdResponseStartAdsVector(builder, len(adOffsets))
	for i := len(adOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(adOffsets[i])
	}
	adsVector := builder.EndVector(len(adOffsets))

	fb.AdResponseStart(builder)
	fb.AdResponseAddAds(builder, adsVector)
	obj := fb.AdResponseEnd(builder)

	builder.Finish(obj)
	return builder.FinishedBytes(), nil
}

// Cap'n Proto converters

func ProtoToCapnp_Empty(pbMsg *pb.Empty) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	_, err = pbcapnp.NewRootEmpty(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Empty: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_EmptyUser(pbMsg *pb.EmptyUser) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	emptyUser, err := pbcapnp.NewRootEmptyUser(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create EmptyUser: %w", err)
	}

	err = emptyUser.SetUserId(pbMsg.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("failed to set user_id: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_Money(pbMsg *pb.Money) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	money, err := pbcapnp.NewRootMoney(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Money: %w", err)
	}

	err = money.SetCurrencyCode(pbMsg.GetCurrencyCode())
	if err != nil {
		return nil, fmt.Errorf("failed to set currency_code: %w", err)
	}
	money.SetUnits(pbMsg.GetUnits())
	money.SetNanos(pbMsg.GetNanos())

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_Address(pbMsg *pb.Address) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	addr, err := pbcapnp.NewRootAddress(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Address: %w", err)
	}

	err = addr.SetStreetAddress(pbMsg.GetStreetAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to set street_address: %w", err)
	}
	err = addr.SetCity(pbMsg.GetCity())
	if err != nil {
		return nil, fmt.Errorf("failed to set city: %w", err)
	}
	err = addr.SetState(pbMsg.GetState())
	if err != nil {
		return nil, fmt.Errorf("failed to set state: %w", err)
	}
	err = addr.SetCountry(pbMsg.GetCountry())
	if err != nil {
		return nil, fmt.Errorf("failed to set country: %w", err)
	}
	addr.SetZipCode(pbMsg.GetZipCode())

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_CartItem(pbMsg *pb.CartItem) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	cartItem, err := pbcapnp.NewRootCartItem(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create CartItem: %w", err)
	}

	err = cartItem.SetProductId(pbMsg.GetProductId())
	if err != nil {
		return nil, fmt.Errorf("failed to set product_id: %w", err)
	}
	cartItem.SetQuantity(pbMsg.GetQuantity())

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_CreditCardInfo(pbMsg *pb.CreditCardInfo) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	cc, err := pbcapnp.NewRootCreditCardInfo(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create CreditCardInfo: %w", err)
	}

	err = cc.SetCreditCardNumber(pbMsg.GetCreditCardNumber())
	if err != nil {
		return nil, fmt.Errorf("failed to set credit_card_number: %w", err)
	}
	cc.SetCreditCardCvv(pbMsg.GetCreditCardCvv())
	cc.SetCreditCardExpirationYear(pbMsg.GetCreditCardExpirationYear())
	cc.SetCreditCardExpirationMonth(pbMsg.GetCreditCardExpirationMonth())

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_Ad(pbMsg *pb.Ad) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	ad, err := pbcapnp.NewRootAd(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Ad: %w", err)
	}

	err = ad.SetRedirectUrl(pbMsg.GetRedirectUrl())
	if err != nil {
		return nil, fmt.Errorf("failed to set redirect_url: %w", err)
	}
	err = ad.SetText(pbMsg.GetText())
	if err != nil {
		return nil, fmt.Errorf("failed to set text: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_AddItemRequest(pbMsg *pb.AddItemRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootAddItemRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create AddItemRequest: %w", err)
	}

	err = req.SetUserId(pbMsg.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("failed to set user_id: %w", err)
	}

	// Create CartItem
	item, err := pbcapnp.NewCartItem(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create CartItem: %w", err)
	}
	err = item.SetProductId(pbMsg.GetItem().GetProductId())
	if err != nil {
		return nil, fmt.Errorf("failed to set product_id: %w", err)
	}
	item.SetQuantity(pbMsg.GetItem().GetQuantity())

	err = req.SetItem(item)
	if err != nil {
		return nil, fmt.Errorf("failed to set item: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_GetCartRequest(pbMsg *pb.GetCartRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootGetCartRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetCartRequest: %w", err)
	}

	err = req.SetUserId(pbMsg.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("failed to set user_id: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_EmptyCartRequest(pbMsg *pb.EmptyCartRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootEmptyCartRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create EmptyCartRequest: %w", err)
	}

	err = req.SetUserId(pbMsg.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("failed to set user_id: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_Cart(pbMsg *pb.Cart) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	cart, err := pbcapnp.NewRootCart(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cart: %w", err)
	}

	err = cart.SetUserId(pbMsg.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("failed to set user_id: %w", err)
	}

	// Create items list
	items := pbMsg.GetItems()
	itemsList, err := pbcapnp.NewCartItem_List(seg, int32(len(items)))
	if err != nil {
		return nil, fmt.Errorf("failed to create items list: %w", err)
	}

	for i, item := range items {
		cartItem := itemsList.At(i)
		err = cartItem.SetProductId(item.GetProductId())
		if err != nil {
			return nil, fmt.Errorf("failed to set product_id for item %d: %w", i, err)
		}
		cartItem.SetQuantity(item.GetQuantity())
	}

	err = cart.SetItems(itemsList)
	if err != nil {
		return nil, fmt.Errorf("failed to set items: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

// Continue with remaining Cap'n Proto converters...
// Due to length, I'll create these in a second file

