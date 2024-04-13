package getter

import (
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/P1coFly/CarInfoEM/http-server/handlers/err_response"
	"github.com/P1coFly/CarInfoEM/internal/models/car"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type GetCar interface {
	GetCars(pageSize, pageToken int, carFilter car.CarFilter) ([]car.CarWithOwner, error)
	GetTotalCarsCount(carFilter car.CarFilter) (int, error)
}

type Info struct {
	Total    int `json:"total"`
	Page     int `json:"page"`
	LastPage int `json:"last_page"`
}

type GetResponse struct {
	CarWithOwner []car.CarWithOwner
	Info         `json:"info"`
}

// @Summary Get
// @Tags cars
// @Description get cars
// @Accept json
// @Produce json
// @Param page_size query int false "Page size (default is 100)" default:"100"
// @Param page_token query int false "Page token (default is 1)" default:"1"
// @Param year query string false "Filter by year (format: 'start:end') example: 2000:2023"
// @Param reg_num query string false "Filter by registration number"
// @Param model query string false "Filter by car model"
// @Param mark query string false "Filter by car mark"
// @Param name query string false "Filter by owner name"
// @Param surname query string false "Filter by owner surname"
// @Param patronymic query string false "Filter by owner patronymic"
// @Success 200 {object} GetResponse
// @Failure 400 {object} err_response.Response
// @Failure 500 {object} err_response.Response
// @Failure default {object} err_response.Response
// @Router /cars [get]
func New(log *slog.Logger, get GetCar) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.GetCar.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// получаемя page_size и page_token для пагинации
		pageSizeStr := r.URL.Query().Get("page_size")
		if pageSizeStr == "" {
			pageSizeStr = "100"
		}
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			log.Error("failed to get page_size", "error", err)
			w.WriteHeader(400)
			render.JSON(w, r, err_response.Error("failed to get page_size. Need page_size=<int>"))

			return
		}
		if pageSize < 1 {
			log.Error("incorrect page_size, page_size must be greater than 0")
			w.WriteHeader(400)
			render.JSON(w, r, err_response.Error("incorrect page_size, page_size must be greater than 0"))

			return
		}

		pageTokenStr := r.URL.Query().Get("page_token")
		if pageTokenStr == "" {
			pageTokenStr = "1"
		}
		pageToken, err := strconv.Atoi(pageTokenStr)
		if err != nil {
			log.Error("failed to get page_token", "error", err)
			w.WriteHeader(400)
			render.JSON(w, r, err_response.Error("failed to get page_token. Need page_token=<int>"))

			return
		}
		if pageToken < 1 {
			log.Error("incorrect page_token, page_token must be greater than 0")
			w.WriteHeader(400)
			render.JSON(w, r, err_response.Error("incorrect page_token, page_token must be greater than 0"))

			return
		}

		// Инициализируем carFilter
		carFilter := car.CarFilter{YearFilter: r.URL.Query().Get("year"),
			RegNumFilter:     r.URL.Query().Get("reg_num"),
			ModelFilter:      r.URL.Query().Get("model"),
			MarkFilter:       r.URL.Query().Get("mark"),
			NameFilter:       r.URL.Query().Get("name"),
			SurnameFilter:    r.URL.Query().Get("surname"),
			PatronymicFilter: r.URL.Query().Get("patronymic")}

		var startYear, endYear int
		if carFilter.YearFilter != "" {
			years := strings.Split(carFilter.YearFilter, ":")
			if len(years) != 2 {
				log.Error("invalid year filter format")
				w.WriteHeader(400)
				render.JSON(w, r, err_response.Error("invalid year filter format"))
				return
			}

			var err error
			startYear, err = strconv.Atoi(years[0])
			if err != nil {
				log.Error("invalid start year", "error", err)
				w.WriteHeader(400)
				render.JSON(w, r, err_response.Error("invalid start year"))
				return
			}

			endYear, err = strconv.Atoi(years[1])
			if err != nil {
				log.Error("invalid end year", "error", err)
				w.WriteHeader(400)
				render.JSON(w, r, err_response.Error("invalid end year"))
				return
			}

			if startYear > endYear {
				log.Error("start year cannot be greater than end year")
				w.WriteHeader(400)
				render.JSON(w, r, err_response.Error("start year cannot be greater than end year"))
				return
			}
		}

		carWithOwner, err := get.GetCars(pageSize, pageToken, carFilter)
		if err != nil {
			log.Error("failed to get cars", "error", err)
			w.WriteHeader(500)
			render.JSON(w, r, err_response.Error("failed to get cars. Try later"))

			return
		}
		log.Info("cars was got")

		total, err := get.GetTotalCarsCount(carFilter)
		if err != nil {
			log.Error("failed to get total cars", "error", err)
			w.WriteHeader(500)
			render.JSON(w, r, err_response.Error("failed to get cars. Try later"))

			return
		}

		log.Info("cars was got")

		w.WriteHeader(200)
		render.JSON(w, r, GetResponse{
			CarWithOwner: carWithOwner,
			Info: Info{
				Total:    total,
				Page:     pageToken,
				LastPage: int(math.Ceil(float64(total) / float64(pageSize)))},
		})

	}
}
