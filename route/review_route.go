package route

import (
	"github.com/gofiber/fiber/v2"

	"flash-sale-api/app/service"
)

// RegisterReviewRoutes mendaftarkan endpoint ulasan (MongoDB).
func RegisterReviewRoutes(api fiber.Router, svc *service.ReviewService) {
	api.Post("/products/:id/reviews", svc.CreateReview)
	api.Get("/products/:id/reviews", svc.GetReviews)
	api.Get("/products/:id/reviews/summary", svc.GetRatingSummary)
	api.Delete("/reviews/:reviewId", svc.DeleteReview)
}
