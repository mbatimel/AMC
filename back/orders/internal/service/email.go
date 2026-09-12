package service

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/google/uuid"

	"github.com/mbatimel/AMC/orders/pkg/models"
)

// sendOrderConfirmationEmail renders and sends the order confirmation email.
// images is keyed by product ID; products with no image are simply skipped.
func (s *service) sendOrderConfirmationEmail(ctx context.Context, to string, order models.Order, images map[uuid.UUID]string) error {
	if s.mailer == nil {
		return nil
	}
	subject := fmt.Sprintf("Заказ №%s оформлен", order.Number)
	return s.mailer.SendHTML(ctx, to, subject, buildOrderConfirmationHTML(order, images))
}

func buildOrderConfirmationHTML(order models.Order, images map[uuid.UUID]string) string {
	var b strings.Builder

	fmt.Fprintf(&b, `<!DOCTYPE html><html><body style="font-family:sans-serif;color:#222;">`)
	fmt.Fprintf(&b, `<h2>Заказ №%s принят</h2>`, html.EscapeString(order.Number))
	fmt.Fprintf(&b, `<p>Здравствуйте, %s! Ваш заказ оформлен, вот его состав:</p>`, html.EscapeString(order.ContactName))

	b.WriteString(`<table style="border-collapse:collapse;width:100%;" cellpadding="8">`)
	b.WriteString(`<thead><tr style="background:#f2f2f2;text-align:left;">`)
	b.WriteString(`<th>Товар</th><th>Артикул</th><th>Кол-во</th><th>Цена</th><th>Сумма</th>`)
	b.WriteString(`</tr></thead><tbody>`)

	for _, item := range order.Items {
		b.WriteString(`<tr style="border-bottom:1px solid #e0e0e0;">`)
		b.WriteString(`<td>`)
		if productID, err := uuid.Parse(item.ProductID); err == nil {
			if imgURL, ok := images[productID]; ok && imgURL != "" {
				fmt.Fprintf(&b, `<img src="%s" alt="%s" width="60" height="60" style="object-fit:cover;vertical-align:middle;margin-right:8px;border-radius:4px;">`,
					html.EscapeString(imgURL), html.EscapeString(item.ProductName))
			}
		}
		fmt.Fprintf(&b, `<span style="vertical-align:middle;">%s</span></td>`, html.EscapeString(item.ProductName))
		fmt.Fprintf(&b, `<td>%s</td>`, html.EscapeString(item.SKU))
		fmt.Fprintf(&b, `<td>%d</td>`, item.Qty)
		fmt.Fprintf(&b, `<td>%s ₸</td>`, formatMoney(item.Price))
		fmt.Fprintf(&b, `<td>%s ₸</td>`, formatMoney(item.Total))
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table>`)

	b.WriteString(`<table style="margin-top:16px;">`)
	fmt.Fprintf(&b, `<tr><td>Сумма без скидки:</td><td style="text-align:right;">%s ₸</td></tr>`, formatMoney(order.Subtotal))
	if order.DiscountTotal > 0 {
		fmt.Fprintf(&b, `<tr><td>Скидка:</td><td style="text-align:right;">-%s ₸</td></tr>`, formatMoney(order.DiscountTotal))
	}
	fmt.Fprintf(&b, `<tr><td>НДС:</td><td style="text-align:right;">%s ₸</td></tr>`, formatMoney(order.VAT))
	fmt.Fprintf(&b, `<tr><td style="font-weight:bold;">Итого к оплате:</td><td style="text-align:right;font-weight:bold;">%s ₸</td></tr>`, formatMoney(order.Total))
	b.WriteString(`</table>`)

	fmt.Fprintf(&b, `<p style="margin-top:16px;">Адрес доставки: %s</p>`, html.EscapeString(order.DeliveryAddress))
	if order.Comment != "" {
		fmt.Fprintf(&b, `<p>Комментарий: %s</p>`, html.EscapeString(order.Comment))
	}

	b.WriteString(`</body></html>`)
	return b.String()
}

func formatMoney(v float64) string {
	return fmt.Sprintf("%.2f", v)
}
