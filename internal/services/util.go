package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/vogiaan1904/order-svc/internal/models"
	order "github.com/vogiaan1904/order-svc/protogen/golang/order"
	"github.com/vogiaan1904/order-svc/protogen/golang/product"
)

func (svc *implOrderService) protoOrderStatusToPtr(status order.OrderStatus) *models.OrderStatus {
	if status == order.OrderStatus_ORDER_STATUS_UNSPECIFIED {
		return nil
	}
	str, ok := order.OrderStatus_name[int32(status)]
	if !ok {
		return nil
	}
	s := models.OrderStatus(str)
	return &s
}

func (svc *implOrderService) validateOrderItems(ctx context.Context, req *order.CreateRequest, resp *product.ListResponse) error {
	if len(resp.Products) != len(req.Items) {
		svc.l.Warnf(ctx, "orderSvc.Create: %v", ErrProductNotFound)
		return ErrProductNotFound
	}

	pMap := make(map[string]*product.ProductData)
	for _, p := range resp.Products {
		pMap[p.Id] = p
	}

	for _, item := range req.Items {
		p, ok := pMap[item.ProductId]
		if !ok {
			svc.l.Warnf(ctx, "orderSvc.Create: Product %v not found", item.ProductId)
			return ErrProductNotFound
		}
		if p.TotalStock < int32(item.Quantity) {
			svc.l.Warnf(ctx, "orderSvc.Create: Product %v does not have enough stock. Requested: %d, Available: %d", item.ProductId, item.Quantity, p.TotalStock)
			return ErrProductOutOfStock
		}
	}

	return nil
}

func (svc *implOrderService) toOrderDataResp(o models.Order) *order.OrderData {
	od := &order.OrderData{
		Id:          o.ID.Hex(),
		UserId:      o.UserID,
		Items:       svc.toOrderItemResp(o.Items),
		TotalAmount: o.TotalAmount,
		Status:      order.OrderStatus(order.OrderStatus_value[string(o.Status)]),
	}

	return od
}

func (svc *implOrderService) toOrderItemResp(items []models.OrderItem) []*order.OrderItem {
	oItems := make([]*order.OrderItem, len(items))
	for i, item := range items {
		oItems[i] = &order.OrderItem{
			ProductId:    item.ProductID,
			Quantity:     item.Quantity,
			ProductPrice: item.ProductPrice,
		}
	}

	return oItems
}

func (svc *implOrderService) generateOrderCode() string {
	now := time.Now()
	// Format date as DDMMYYYY
	dateStr := now.Format("02012006")

	// Generate 8 random digits
	randomNum := rand.Intn(100000000)
	// Ensure 8 digits by padding with leading zeros if necessary
	randomStr := fmt.Sprintf("%08d", randomNum)

	// Combine date and random number
	return dateStr + randomStr
}
