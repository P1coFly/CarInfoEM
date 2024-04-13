package patcher

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/P1coFly/CarInfoEM/http-server/handlers/err_response"
	"github.com/P1coFly/CarInfoEM/internal/models/car"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	car.PatchCar
}

type PatcherCar interface {
	PatchCar(carID int, cwo car.PatchCar) (int, error)
}

// @Summary Patch
// @Tags car
// @Description patch car
// @Accept json
// @Produce json
// @Param id path int true "Car ID"
// @Param input body car.Car true "new car data"
// @Success 200
// @Failure 400,404 {object} err_response.Response
// @Failure 500 {object} err_response.Response
// @Failure default {object} err_response.Response
// @Router /car/patch/{id} [patch]
func New(log *slog.Logger, patcher PatcherCar) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.PatcherCar.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		//пытаемсяя получить id с запроса
		carID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Error("failed to get car ID from URL")
			w.WriteHeader(400)
			render.JSON(w, r, err_response.Error("failed to get car ID from URL"))
			return
		}

		//декодируем тело запроса
		var req Request
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode request body", "error", err)
			w.WriteHeader(400)
			render.JSON(w, r, err_response.Error("failed to decode request body"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		//вызываем метож патча сущности
		code, err := patcher.PatchCar(carID, req.PatchCar)
		if err != nil {
			if code == -1 {
				w.WriteHeader(404)
				render.JSON(w, r, err_response.Error("car with this id was not found"))
				return
			}
			log.Error("failed to patch car", "error", err)
			w.WriteHeader(500)
			render.JSON(w, r, err_response.Error(fmt.Sprintf("failed to patch car: %v", err)))
			return
		}
		//Если нет изменений
		if code == 2 {
			w.WriteHeader(204)
			return
		}

		w.WriteHeader(200)
	}
}
