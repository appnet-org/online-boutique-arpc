package messagelogger

import (
	"fmt"

	pb "github.com/appnetorg/online-boutique-arpc/proto"
	pbcapnp "github.com/appnetorg/online-boutique-arpc/proto/capnp"
	capnp "capnproto.org/go/capnp/v3"
)

// Remaining Cap'n Proto converters

func ProtoToCapnp_ListRecommendationsRequest(pbMsg *pb.ListRecommendationsRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootListRecommendationsRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ListRecommendationsRequest: %w", err)
	}

	err = req.SetUserId(pbMsg.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("failed to set user_id: %w", err)
	}

	// Create product IDs list
	productIDs := pbMsg.GetProductIds()
	productIDsList, err := capnp.NewTextList(seg, int32(len(productIDs)))
	if err != nil {
		return nil, fmt.Errorf("failed to create product_ids list: %w", err)
	}

	for i, pid := range productIDs {
		err = productIDsList.Set(i, pid)
		if err != nil {
			return nil, fmt.Errorf("failed to set product_id %d: %w", i, err)
		}
	}

	err = req.SetProductIds(productIDsList)
	if err != nil {
		return nil, fmt.Errorf("failed to set product_ids: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_ListRecommendationsResponse(pbMsg *pb.ListRecommendationsResponse) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	resp, err := pbcapnp.NewRootListRecommendationsResponse(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ListRecommendationsResponse: %w", err)
	}

	// Create product IDs list
	productIDs := pbMsg.GetProductIds()
	productIDsList, err := capnp.NewTextList(seg, int32(len(productIDs)))
	if err != nil {
		return nil, fmt.Errorf("failed to create product_ids list: %w", err)
	}

	for i, pid := range productIDs {
		err = productIDsList.Set(i, pid)
		if err != nil {
			return nil, fmt.Errorf("failed to set product_id %d: %w", i, err)
		}
	}

	err = resp.SetProductIds(productIDsList)
	if err != nil {
		return nil, fmt.Errorf("failed to set product_ids: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_Product(pbMsg *pb.Product) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	product, err := pbcapnp.NewRootProduct(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Product: %w", err)
	}

	err = product.SetId(pbMsg.GetId())
	if err != nil {
		return nil, fmt.Errorf("failed to set id: %w", err)
	}
	err = product.SetName(pbMsg.GetName())
	if err != nil {
		return nil, fmt.Errorf("failed to set name: %w", err)
	}
	err = product.SetDescription(pbMsg.GetDescription())
	if err != nil {
		return nil, fmt.Errorf("failed to set description: %w", err)
	}
	err = product.SetPicture(pbMsg.GetPicture())
	if err != nil {
		return nil, fmt.Errorf("failed to set picture: %w", err)
	}

	// Create Money
	priceUsd := pbMsg.GetPriceUsd()
	money, err := pbcapnp.NewMoney(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Money: %w", err)
	}
	err = money.SetCurrencyCode(priceUsd.GetCurrencyCode())
	if err != nil {
		return nil, fmt.Errorf("failed to set currency_code: %w", err)
	}
	money.SetUnits(priceUsd.GetUnits())
	money.SetNanos(priceUsd.GetNanos())

	err = product.SetPriceUsd(money)
	if err != nil {
		return nil, fmt.Errorf("failed to set price_usd: %w", err)
	}

	// Create categories list
	categories := pbMsg.GetCategories()
	categoriesList, err := capnp.NewTextList(seg, int32(len(categories)))
	if err != nil {
		return nil, fmt.Errorf("failed to create categories list: %w", err)
	}

	for i, cat := range categories {
		err = categoriesList.Set(i, cat)
		if err != nil {
			return nil, fmt.Errorf("failed to set category %d: %w", i, err)
		}
	}

	err = product.SetCategories(categoriesList)
	if err != nil {
		return nil, fmt.Errorf("failed to set categories: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_GetProductRequest(pbMsg *pb.GetProductRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootGetProductRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetProductRequest: %w", err)
	}

	err = req.SetId(pbMsg.GetId())
	if err != nil {
		return nil, fmt.Errorf("failed to set id: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_SearchProductsRequest(pbMsg *pb.SearchProductsRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootSearchProductsRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create SearchProductsRequest: %w", err)
	}

	err = req.SetQuery(pbMsg.GetQuery())
	if err != nil {
		return nil, fmt.Errorf("failed to set query: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_ListProductsResponse(pbMsg *pb.ListProductsResponse) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	resp, err := pbcapnp.NewRootListProductsResponse(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ListProductsResponse: %w", err)
	}

	// Create products list
	products := pbMsg.GetProducts()
	productsList, err := pbcapnp.NewProduct_List(seg, int32(len(products)))
	if err != nil {
		return nil, fmt.Errorf("failed to create products list: %w", err)
	}

	for i, prod := range products {
		product := productsList.At(i)

		err = product.SetId(prod.GetId())
		if err != nil {
			return nil, fmt.Errorf("failed to set id for product %d: %w", i, err)
		}
		err = product.SetName(prod.GetName())
		if err != nil {
			return nil, fmt.Errorf("failed to set name for product %d: %w", i, err)
		}
		err = product.SetDescription(prod.GetDescription())
		if err != nil {
			return nil, fmt.Errorf("failed to set description for product %d: %w", i, err)
		}
		err = product.SetPicture(prod.GetPicture())
		if err != nil {
			return nil, fmt.Errorf("failed to set picture for product %d: %w", i, err)
		}

		// Create Money
		priceUsd := prod.GetPriceUsd()
		money, err := pbcapnp.NewMoney(seg)
		if err != nil {
			return nil, fmt.Errorf("failed to create Money for product %d: %w", i, err)
		}
		err = money.SetCurrencyCode(priceUsd.GetCurrencyCode())
		if err != nil {
			return nil, fmt.Errorf("failed to set currency_code for product %d: %w", i, err)
		}
		money.SetUnits(priceUsd.GetUnits())
		money.SetNanos(priceUsd.GetNanos())

		err = product.SetPriceUsd(money)
		if err != nil {
			return nil, fmt.Errorf("failed to set price_usd for product %d: %w", i, err)
		}

		// Create categories list
		categories := prod.GetCategories()
		categoriesList, err := capnp.NewTextList(seg, int32(len(categories)))
		if err != nil {
			return nil, fmt.Errorf("failed to create categories list for product %d: %w", i, err)
		}

		for j, cat := range categories {
			err = categoriesList.Set(j, cat)
			if err != nil {
				return nil, fmt.Errorf("failed to set category %d for product %d: %w", j, i, err)
			}
		}

		err = product.SetCategories(categoriesList)
		if err != nil {
			return nil, fmt.Errorf("failed to set categories for product %d: %w", i, err)
		}
	}

	err = resp.SetProducts(productsList)
	if err != nil {
		return nil, fmt.Errorf("failed to set products: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_SearchProductsResponse(pbMsg *pb.SearchProductsResponse) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	resp, err := pbcapnp.NewRootSearchProductsResponse(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create SearchProductsResponse: %w", err)
	}

	// Create results list (same as products)
	results := pbMsg.GetResults()
	resultsList, err := pbcapnp.NewProduct_List(seg, int32(len(results)))
	if err != nil {
		return nil, fmt.Errorf("failed to create results list: %w", err)
	}

	for i, prod := range results {
		product := resultsList.At(i)

		err = product.SetId(prod.GetId())
		if err != nil {
			return nil, fmt.Errorf("failed to set id for result %d: %w", i, err)
		}
		err = product.SetName(prod.GetName())
		if err != nil {
			return nil, fmt.Errorf("failed to set name for result %d: %w", i, err)
		}
		err = product.SetDescription(prod.GetDescription())
		if err != nil {
			return nil, fmt.Errorf("failed to set description for result %d: %w", i, err)
		}
		err = product.SetPicture(prod.GetPicture())
		if err != nil {
			return nil, fmt.Errorf("failed to set picture for result %d: %w", i, err)
		}

		// Create Money
		priceUsd := prod.GetPriceUsd()
		money, err := pbcapnp.NewMoney(seg)
		if err != nil {
			return nil, fmt.Errorf("failed to create Money for result %d: %w", i, err)
		}
		err = money.SetCurrencyCode(priceUsd.GetCurrencyCode())
		if err != nil {
			return nil, fmt.Errorf("failed to set currency_code for result %d: %w", i, err)
		}
		money.SetUnits(priceUsd.GetUnits())
		money.SetNanos(priceUsd.GetNanos())

		err = product.SetPriceUsd(money)
		if err != nil {
			return nil, fmt.Errorf("failed to set price_usd for result %d: %w", i, err)
		}

		// Create categories list
		categories := prod.GetCategories()
		categoriesList, err := capnp.NewTextList(seg, int32(len(categories)))
		if err != nil {
			return nil, fmt.Errorf("failed to create categories list for result %d: %w", i, err)
		}

		for j, cat := range categories {
			err = categoriesList.Set(j, cat)
			if err != nil {
				return nil, fmt.Errorf("failed to set category %d for result %d: %w", j, i, err)
			}
		}

		err = product.SetCategories(categoriesList)
		if err != nil {
			return nil, fmt.Errorf("failed to set categories for result %d: %w", i, err)
		}
	}

	err = resp.SetResults(resultsList)
	if err != nil {
		return nil, fmt.Errorf("failed to set results: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_GetQuoteRequest(pbMsg *pb.GetQuoteRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootGetQuoteRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetQuoteRequest: %w", err)
	}

	// Create Address
	addr := pbMsg.GetAddress()
	address, err := pbcapnp.NewAddress(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Address: %w", err)
	}
	err = address.SetStreetAddress(addr.GetStreetAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to set street_address: %w", err)
	}
	err = address.SetCity(addr.GetCity())
	if err != nil {
		return nil, fmt.Errorf("failed to set city: %w", err)
	}
	err = address.SetState(addr.GetState())
	if err != nil {
		return nil, fmt.Errorf("failed to set state: %w", err)
	}
	err = address.SetCountry(addr.GetCountry())
	if err != nil {
		return nil, fmt.Errorf("failed to set country: %w", err)
	}
	address.SetZipCode(addr.GetZipCode())

	err = req.SetAddress(address)
	if err != nil {
		return nil, fmt.Errorf("failed to set address: %w", err)
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

	err = req.SetItems(itemsList)
	if err != nil {
		return nil, fmt.Errorf("failed to set items: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_GetQuoteResponse(pbMsg *pb.GetQuoteResponse) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	resp, err := pbcapnp.NewRootGetQuoteResponse(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetQuoteResponse: %w", err)
	}

	// Create Money
	costUsd := pbMsg.GetCostUsd()
	money, err := pbcapnp.NewMoney(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Money: %w", err)
	}
	err = money.SetCurrencyCode(costUsd.GetCurrencyCode())
	if err != nil {
		return nil, fmt.Errorf("failed to set currency_code: %w", err)
	}
	money.SetUnits(costUsd.GetUnits())
	money.SetNanos(costUsd.GetNanos())

	err = resp.SetCostUsd(money)
	if err != nil {
		return nil, fmt.Errorf("failed to set cost_usd: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_ShipOrderRequest(pbMsg *pb.ShipOrderRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootShipOrderRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ShipOrderRequest: %w", err)
	}

	// Create Address
	addr := pbMsg.GetAddress()
	address, err := pbcapnp.NewAddress(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Address: %w", err)
	}
	err = address.SetStreetAddress(addr.GetStreetAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to set street_address: %w", err)
	}
	err = address.SetCity(addr.GetCity())
	if err != nil {
		return nil, fmt.Errorf("failed to set city: %w", err)
	}
	err = address.SetState(addr.GetState())
	if err != nil {
		return nil, fmt.Errorf("failed to set state: %w", err)
	}
	err = address.SetCountry(addr.GetCountry())
	if err != nil {
		return nil, fmt.Errorf("failed to set country: %w", err)
	}
	address.SetZipCode(addr.GetZipCode())

	err = req.SetAddress(address)
	if err != nil {
		return nil, fmt.Errorf("failed to set address: %w", err)
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

	err = req.SetItems(itemsList)
	if err != nil {
		return nil, fmt.Errorf("failed to set items: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_ShipOrderResponse(pbMsg *pb.ShipOrderResponse) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	resp, err := pbcapnp.NewRootShipOrderResponse(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ShipOrderResponse: %w", err)
	}

	err = resp.SetTrackingId(pbMsg.GetTrackingId())
	if err != nil {
		return nil, fmt.Errorf("failed to set tracking_id: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_GetSupportedCurrenciesResponse(pbMsg *pb.GetSupportedCurrenciesResponse) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	resp, err := pbcapnp.NewRootGetSupportedCurrenciesResponse(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetSupportedCurrenciesResponse: %w", err)
	}

	// Create currency codes list
	currencyCodes := pbMsg.GetCurrencyCodes()
	codesList, err := capnp.NewTextList(seg, int32(len(currencyCodes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create currency_codes list: %w", err)
	}

	for i, code := range currencyCodes {
		err = codesList.Set(i, code)
		if err != nil {
			return nil, fmt.Errorf("failed to set currency_code %d: %w", i, err)
		}
	}

	err = resp.SetCurrencyCodes(codesList)
	if err != nil {
		return nil, fmt.Errorf("failed to set currency_codes: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_CurrencyConversionRequest(pbMsg *pb.CurrencyConversionRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootCurrencyConversionRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create CurrencyConversionRequest: %w", err)
	}

	// Create Money
	from := pbMsg.GetFrom()
	money, err := pbcapnp.NewMoney(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Money: %w", err)
	}
	err = money.SetCurrencyCode(from.GetCurrencyCode())
	if err != nil {
		return nil, fmt.Errorf("failed to set currency_code: %w", err)
	}
	money.SetUnits(from.GetUnits())
	money.SetNanos(from.GetNanos())

	err = req.SetFrom(money)
	if err != nil {
		return nil, fmt.Errorf("failed to set from: %w", err)
	}

	err = req.SetToCode(pbMsg.GetToCode())
	if err != nil {
		return nil, fmt.Errorf("failed to set to_code: %w", err)
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

func ProtoToCapnp_ChargeRequest(pbMsg *pb.ChargeRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootChargeRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ChargeRequest: %w", err)
	}

	// Create Money
	amount := pbMsg.GetAmount()
	money, err := pbcapnp.NewMoney(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Money: %w", err)
	}
	err = money.SetCurrencyCode(amount.GetCurrencyCode())
	if err != nil {
		return nil, fmt.Errorf("failed to set currency_code: %w", err)
	}
	money.SetUnits(amount.GetUnits())
	money.SetNanos(amount.GetNanos())

	err = req.SetAmount(money)
	if err != nil {
		return nil, fmt.Errorf("failed to set amount: %w", err)
	}

	// Create CreditCardInfo
	cc := pbMsg.GetCreditCard()
	creditCard, err := pbcapnp.NewCreditCardInfo(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create CreditCardInfo: %w", err)
	}
	err = creditCard.SetCreditCardNumber(cc.GetCreditCardNumber())
	if err != nil {
		return nil, fmt.Errorf("failed to set credit_card_number: %w", err)
	}
	creditCard.SetCreditCardCvv(cc.GetCreditCardCvv())
	creditCard.SetCreditCardExpirationYear(cc.GetCreditCardExpirationYear())
	creditCard.SetCreditCardExpirationMonth(cc.GetCreditCardExpirationMonth())

	err = req.SetCreditCard(creditCard)
	if err != nil {
		return nil, fmt.Errorf("failed to set credit_card: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_ChargeResponse(pbMsg *pb.ChargeResponse) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	resp, err := pbcapnp.NewRootChargeResponse(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ChargeResponse: %w", err)
	}

	err = resp.SetTransactionId(pbMsg.GetTransactionId())
	if err != nil {
		return nil, fmt.Errorf("failed to set transaction_id: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_OrderItem(pbMsg *pb.OrderItem) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	orderItem, err := pbcapnp.NewRootOrderItem(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OrderItem: %w", err)
	}

	// Create CartItem
	item := pbMsg.GetItem()
	cartItem, err := pbcapnp.NewCartItem(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create CartItem: %w", err)
	}
	err = cartItem.SetProductId(item.GetProductId())
	if err != nil {
		return nil, fmt.Errorf("failed to set product_id: %w", err)
	}
	cartItem.SetQuantity(item.GetQuantity())

	err = orderItem.SetItem(cartItem)
	if err != nil {
		return nil, fmt.Errorf("failed to set item: %w", err)
	}

	// Create Money
	cost := pbMsg.GetCost()
	money, err := pbcapnp.NewMoney(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Money: %w", err)
	}
	err = money.SetCurrencyCode(cost.GetCurrencyCode())
	if err != nil {
		return nil, fmt.Errorf("failed to set currency_code: %w", err)
	}
	money.SetUnits(cost.GetUnits())
	money.SetNanos(cost.GetNanos())

	err = orderItem.SetCost(money)
	if err != nil {
		return nil, fmt.Errorf("failed to set cost: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_OrderResult(pbMsg *pb.OrderResult) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	orderResult, err := pbcapnp.NewRootOrderResult(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OrderResult: %w", err)
	}

	err = orderResult.SetOrderId(pbMsg.GetOrderId())
	if err != nil {
		return nil, fmt.Errorf("failed to set order_id: %w", err)
	}

	err = orderResult.SetShippingTrackingId(pbMsg.GetShippingTrackingId())
	if err != nil {
		return nil, fmt.Errorf("failed to set shipping_tracking_id: %w", err)
	}

	// Create Money (shipping cost)
	shippingCost := pbMsg.GetShippingCost()
	money, err := pbcapnp.NewMoney(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Money: %w", err)
	}
	err = money.SetCurrencyCode(shippingCost.GetCurrencyCode())
	if err != nil {
		return nil, fmt.Errorf("failed to set currency_code: %w", err)
	}
	money.SetUnits(shippingCost.GetUnits())
	money.SetNanos(shippingCost.GetNanos())

	err = orderResult.SetShippingCost(money)
	if err != nil {
		return nil, fmt.Errorf("failed to set shipping_cost: %w", err)
	}

	// Create Address
	addr := pbMsg.GetShippingAddress()
	address, err := pbcapnp.NewAddress(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Address: %w", err)
	}
	err = address.SetStreetAddress(addr.GetStreetAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to set street_address: %w", err)
	}
	err = address.SetCity(addr.GetCity())
	if err != nil {
		return nil, fmt.Errorf("failed to set city: %w", err)
	}
	err = address.SetState(addr.GetState())
	if err != nil {
		return nil, fmt.Errorf("failed to set state: %w", err)
	}
	err = address.SetCountry(addr.GetCountry())
	if err != nil {
		return nil, fmt.Errorf("failed to set country: %w", err)
	}
	address.SetZipCode(addr.GetZipCode())

	err = orderResult.SetShippingAddress(address)
	if err != nil {
		return nil, fmt.Errorf("failed to set shipping_address: %w", err)
	}

	// Create items list
	items := pbMsg.GetItems()
	itemsList, err := pbcapnp.NewOrderItem_List(seg, int32(len(items)))
	if err != nil {
		return nil, fmt.Errorf("failed to create items list: %w", err)
	}

	for i, item := range items {
		orderItem := itemsList.At(i)

		// Create CartItem
		cartItemPb := item.GetItem()
		cartItem, err := pbcapnp.NewCartItem(seg)
		if err != nil {
			return nil, fmt.Errorf("failed to create CartItem for item %d: %w", i, err)
		}
		err = cartItem.SetProductId(cartItemPb.GetProductId())
		if err != nil {
			return nil, fmt.Errorf("failed to set product_id for item %d: %w", i, err)
		}
		cartItem.SetQuantity(cartItemPb.GetQuantity())

		err = orderItem.SetItem(cartItem)
		if err != nil {
			return nil, fmt.Errorf("failed to set item for order_item %d: %w", i, err)
		}

		// Create Money
		costPb := item.GetCost()
		costMoney, err := pbcapnp.NewMoney(seg)
		if err != nil {
			return nil, fmt.Errorf("failed to create Money for item %d: %w", i, err)
		}
		err = costMoney.SetCurrencyCode(costPb.GetCurrencyCode())
		if err != nil {
			return nil, fmt.Errorf("failed to set currency_code for item %d: %w", i, err)
		}
		costMoney.SetUnits(costPb.GetUnits())
		costMoney.SetNanos(costPb.GetNanos())

		err = orderItem.SetCost(costMoney)
		if err != nil {
			return nil, fmt.Errorf("failed to set cost for order_item %d: %w", i, err)
		}
	}

	err = orderResult.SetItems(itemsList)
	if err != nil {
		return nil, fmt.Errorf("failed to set items: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_SendOrderConfirmationRequest(pbMsg *pb.SendOrderConfirmationRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootSendOrderConfirmationRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create SendOrderConfirmationRequest: %w", err)
	}

	err = req.SetEmail(pbMsg.GetEmail())
	if err != nil {
		return nil, fmt.Errorf("failed to set email: %w", err)
	}

	// Create OrderResult - same logic as ProtoToCapnp_OrderResult but not as root
	order := pbMsg.GetOrder()
	orderResult, err := pbcapnp.NewOrderResult(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OrderResult: %w", err)
	}

	err = orderResult.SetOrderId(order.GetOrderId())
	if err != nil {
		return nil, fmt.Errorf("failed to set order_id: %w", err)
	}

	err = orderResult.SetShippingTrackingId(order.GetShippingTrackingId())
	if err != nil {
		return nil, fmt.Errorf("failed to set shipping_tracking_id: %w", err)
	}

	// Create Money (shipping cost)
	shippingCost := order.GetShippingCost()
	money, err := pbcapnp.NewMoney(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Money: %w", err)
	}
	err = money.SetCurrencyCode(shippingCost.GetCurrencyCode())
	if err != nil {
		return nil, fmt.Errorf("failed to set currency_code: %w", err)
	}
	money.SetUnits(shippingCost.GetUnits())
	money.SetNanos(shippingCost.GetNanos())

	err = orderResult.SetShippingCost(money)
	if err != nil {
		return nil, fmt.Errorf("failed to set shipping_cost: %w", err)
	}

	// Create Address
	addr := order.GetShippingAddress()
	address, err := pbcapnp.NewAddress(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Address: %w", err)
	}
	err = address.SetStreetAddress(addr.GetStreetAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to set street_address: %w", err)
	}
	err = address.SetCity(addr.GetCity())
	if err != nil {
		return nil, fmt.Errorf("failed to set city: %w", err)
	}
	err = address.SetState(addr.GetState())
	if err != nil {
		return nil, fmt.Errorf("failed to set state: %w", err)
	}
	err = address.SetCountry(addr.GetCountry())
	if err != nil {
		return nil, fmt.Errorf("failed to set country: %w", err)
	}
	address.SetZipCode(addr.GetZipCode())

	err = orderResult.SetShippingAddress(address)
	if err != nil {
		return nil, fmt.Errorf("failed to set shipping_address: %w", err)
	}

	// Create items list
	items := order.GetItems()
	itemsList, err := pbcapnp.NewOrderItem_List(seg, int32(len(items)))
	if err != nil {
		return nil, fmt.Errorf("failed to create items list: %w", err)
	}

	for i, item := range items {
		orderItem := itemsList.At(i)

		// Create CartItem
		cartItemPb := item.GetItem()
		cartItem, err := pbcapnp.NewCartItem(seg)
		if err != nil {
			return nil, fmt.Errorf("failed to create CartItem for item %d: %w", i, err)
		}
		err = cartItem.SetProductId(cartItemPb.GetProductId())
		if err != nil {
			return nil, fmt.Errorf("failed to set product_id for item %d: %w", i, err)
		}
		cartItem.SetQuantity(cartItemPb.GetQuantity())

		err = orderItem.SetItem(cartItem)
		if err != nil {
			return nil, fmt.Errorf("failed to set item for order_item %d: %w", i, err)
		}

		// Create Money
		costPb := item.GetCost()
		costMoney, err := pbcapnp.NewMoney(seg)
		if err != nil {
			return nil, fmt.Errorf("failed to create Money for item %d: %w", i, err)
		}
		err = costMoney.SetCurrencyCode(costPb.GetCurrencyCode())
		if err != nil {
			return nil, fmt.Errorf("failed to set currency_code for item %d: %w", i, err)
		}
		costMoney.SetUnits(costPb.GetUnits())
		costMoney.SetNanos(costPb.GetNanos())

		err = orderItem.SetCost(costMoney)
		if err != nil {
			return nil, fmt.Errorf("failed to set cost for order_item %d: %w", i, err)
		}
	}

	err = orderResult.SetItems(itemsList)
	if err != nil {
		return nil, fmt.Errorf("failed to set items: %w", err)
	}

	err = req.SetOrder(orderResult)
	if err != nil {
		return nil, fmt.Errorf("failed to set order: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_PlaceOrderRequest(pbMsg *pb.PlaceOrderRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootPlaceOrderRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create PlaceOrderRequest: %w", err)
	}

	err = req.SetUserId(pbMsg.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("failed to set user_id: %w", err)
	}

	err = req.SetUserCurrency(pbMsg.GetUserCurrency())
	if err != nil {
		return nil, fmt.Errorf("failed to set user_currency: %w", err)
	}

	err = req.SetEmail(pbMsg.GetEmail())
	if err != nil {
		return nil, fmt.Errorf("failed to set email: %w", err)
	}

	// Create Address
	addr := pbMsg.GetAddress()
	address, err := pbcapnp.NewAddress(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Address: %w", err)
	}
	err = address.SetStreetAddress(addr.GetStreetAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to set street_address: %w", err)
	}
	err = address.SetCity(addr.GetCity())
	if err != nil {
		return nil, fmt.Errorf("failed to set city: %w", err)
	}
	err = address.SetState(addr.GetState())
	if err != nil {
		return nil, fmt.Errorf("failed to set state: %w", err)
	}
	err = address.SetCountry(addr.GetCountry())
	if err != nil {
		return nil, fmt.Errorf("failed to set country: %w", err)
	}
	address.SetZipCode(addr.GetZipCode())

	err = req.SetAddress(address)
	if err != nil {
		return nil, fmt.Errorf("failed to set address: %w", err)
	}

	// Create CreditCardInfo
	cc := pbMsg.GetCreditCard()
	creditCard, err := pbcapnp.NewCreditCardInfo(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create CreditCardInfo: %w", err)
	}
	err = creditCard.SetCreditCardNumber(cc.GetCreditCardNumber())
	if err != nil {
		return nil, fmt.Errorf("failed to set credit_card_number: %w", err)
	}
	creditCard.SetCreditCardCvv(cc.GetCreditCardCvv())
	creditCard.SetCreditCardExpirationYear(cc.GetCreditCardExpirationYear())
	creditCard.SetCreditCardExpirationMonth(cc.GetCreditCardExpirationMonth())

	err = req.SetCreditCard(creditCard)
	if err != nil {
		return nil, fmt.Errorf("failed to set credit_card: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_PlaceOrderResponse(pbMsg *pb.PlaceOrderResponse) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	resp, err := pbcapnp.NewRootPlaceOrderResponse(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create PlaceOrderResponse: %w", err)
	}

	// Create OrderResult
	order := pbMsg.GetOrder()
	orderResult, err := pbcapnp.NewOrderResult(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OrderResult: %w", err)
	}

	err = orderResult.SetOrderId(order.GetOrderId())
	if err != nil {
		return nil, fmt.Errorf("failed to set order_id: %w", err)
	}

	err = orderResult.SetShippingTrackingId(order.GetShippingTrackingId())
	if err != nil {
		return nil, fmt.Errorf("failed to set shipping_tracking_id: %w", err)
	}

	// Create Money (shipping cost)
	shippingCost := order.GetShippingCost()
	money, err := pbcapnp.NewMoney(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Money: %w", err)
	}
	err = money.SetCurrencyCode(shippingCost.GetCurrencyCode())
	if err != nil {
		return nil, fmt.Errorf("failed to set currency_code: %w", err)
	}
	money.SetUnits(shippingCost.GetUnits())
	money.SetNanos(shippingCost.GetNanos())

	err = orderResult.SetShippingCost(money)
	if err != nil {
		return nil, fmt.Errorf("failed to set shipping_cost: %w", err)
	}

	// Create Address
	addr := order.GetShippingAddress()
	address, err := pbcapnp.NewAddress(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Address: %w", err)
	}
	err = address.SetStreetAddress(addr.GetStreetAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to set street_address: %w", err)
	}
	err = address.SetCity(addr.GetCity())
	if err != nil {
		return nil, fmt.Errorf("failed to set city: %w", err)
	}
	err = address.SetState(addr.GetState())
	if err != nil {
		return nil, fmt.Errorf("failed to set state: %w", err)
	}
	err = address.SetCountry(addr.GetCountry())
	if err != nil {
		return nil, fmt.Errorf("failed to set country: %w", err)
	}
	address.SetZipCode(addr.GetZipCode())

	err = orderResult.SetShippingAddress(address)
	if err != nil {
		return nil, fmt.Errorf("failed to set shipping_address: %w", err)
	}

	// Create items list
	items := order.GetItems()
	itemsList, err := pbcapnp.NewOrderItem_List(seg, int32(len(items)))
	if err != nil {
		return nil, fmt.Errorf("failed to create items list: %w", err)
	}

	for i, item := range items {
		orderItem := itemsList.At(i)

		// Create CartItem
		cartItemPb := item.GetItem()
		cartItem, err := pbcapnp.NewCartItem(seg)
		if err != nil {
			return nil, fmt.Errorf("failed to create CartItem for item %d: %w", i, err)
		}
		err = cartItem.SetProductId(cartItemPb.GetProductId())
		if err != nil {
			return nil, fmt.Errorf("failed to set product_id for item %d: %w", i, err)
		}
		cartItem.SetQuantity(cartItemPb.GetQuantity())

		err = orderItem.SetItem(cartItem)
		if err != nil {
			return nil, fmt.Errorf("failed to set item for order_item %d: %w", i, err)
		}

		// Create Money
		costPb := item.GetCost()
		costMoney, err := pbcapnp.NewMoney(seg)
		if err != nil {
			return nil, fmt.Errorf("failed to create Money for item %d: %w", i, err)
		}
		err = costMoney.SetCurrencyCode(costPb.GetCurrencyCode())
		if err != nil {
			return nil, fmt.Errorf("failed to set currency_code for item %d: %w", i, err)
		}
		costMoney.SetUnits(costPb.GetUnits())
		costMoney.SetNanos(costPb.GetNanos())

		err = orderItem.SetCost(costMoney)
		if err != nil {
			return nil, fmt.Errorf("failed to set cost for order_item %d: %w", i, err)
		}
	}

	err = orderResult.SetItems(itemsList)
	if err != nil {
		return nil, fmt.Errorf("failed to set items: %w", err)
	}

	err = resp.SetOrder(orderResult)
	if err != nil {
		return nil, fmt.Errorf("failed to set order: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_AdRequest(pbMsg *pb.AdRequest) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	req, err := pbcapnp.NewRootAdRequest(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create AdRequest: %w", err)
	}

	err = req.SetUserId(pbMsg.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("failed to set user_id: %w", err)
	}

	// Create context keys list
	contextKeys := pbMsg.GetContextKeys()
	keysList, err := capnp.NewTextList(seg, int32(len(contextKeys)))
	if err != nil {
		return nil, fmt.Errorf("failed to create context_keys list: %w", err)
	}

	for i, key := range contextKeys {
		err = keysList.Set(i, key)
		if err != nil {
			return nil, fmt.Errorf("failed to set context_key %d: %w", i, err)
		}
	}

	err = req.SetContextKeys(keysList)
	if err != nil {
		return nil, fmt.Errorf("failed to set context_keys: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

func ProtoToCapnp_AdResponse(pbMsg *pb.AdResponse) ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	resp, err := pbcapnp.NewRootAdResponse(seg)
	if err != nil {
		return nil, fmt.Errorf("failed to create AdResponse: %w", err)
	}

	// Create ads list
	ads := pbMsg.GetAds()
	adsList, err := pbcapnp.NewAd_List(seg, int32(len(ads)))
	if err != nil {
		return nil, fmt.Errorf("failed to create ads list: %w", err)
	}

	for i, ad := range ads {
		adItem := adsList.At(i)

		err = adItem.SetRedirectUrl(ad.GetRedirectUrl())
		if err != nil {
			return nil, fmt.Errorf("failed to set redirect_url for ad %d: %w", i, err)
		}

		err = adItem.SetText(ad.GetText())
		if err != nil {
			return nil, fmt.Errorf("failed to set text for ad %d: %w", i, err)
		}
	}

	err = resp.SetAds(adsList)
	if err != nil {
		return nil, fmt.Errorf("failed to set ads: %w", err)
	}

	data, err := msg.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}
