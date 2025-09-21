package handlers

import (
	"bytes"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/MKhiriev/stunning-adventure/internal/utils"
	"github.com/rs/zerolog"
)

func (h *Handler) WithHashing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Получаем запрос и обрабатываем его
		supportsHashing := h.hasher != nil

		// если не поддерживаем хэширование - ничего не делаем с запросом и пропускаем данный handler
		if !supportsHashing {
			h.logger.Debug().
				Str("func", "*Handler.WithHashing").
				Msg("skipping hashing middleware...")
			next.ServeHTTP(w, r)
			return
		}

		// если тело запроса пустое - то нет смысла
		if r.Body == nil {
			h.logger.Debug().
				Str("func", "*Handler.WithHashing").
				Msg("empty body was passed")
			next.ServeHTTP(w, r)
			return
		}

		// далее что делаем?
		// ну, если ключ есть и есть хэшер
		// получаем хэш из заголовка
		hashFromHeader := r.Header.Get("HashSHA256")
		if hashFromHeader == "" {
			h.logger.Debug().
				Str("func", "*Handler.WithHashing").
				Str("hash from header", hashFromHeader).Msg("empty header was passed")
			next.ServeHTTP(w, r)
			return
		}
		h.logger.Debug().
			Str("func", "*Handler.WithHashing").
			Str("hash from header", hashFromHeader).Msg("")

		// что мы делаем с заголовком?
		// сравниваем с телом из запроса
		// тогда первое, что нужно сделать? - получить тело запроса!
		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Err(err).Str("func", "*Handler.WithHashing").Msg("error during reading a body")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		h.logger.Debug().Str("func", "*Handler.WithHashing").Bytes("http body", body).Msg("")

		// return body to the request
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		// получаем хэш тела
		hashedBody, err := h.hasher.HashByteSlice(body)
		if err != nil {
			h.logger.Err(err).Str("func", "*Handler.WithHashing").Msg("error during hashing a body")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		h.logger.Debug().
			Str("func", "*Handler.WithHashing").
			Str("hash from header", hashFromHeader).
			Str("hashed body", hex.EncodeToString(hashedBody)).
			Msg("")

		// сравниваем хэши

		// если хэши не совпадают - ошибка
		if hashFromHeader != hex.EncodeToString(hashedBody) {
			// возвращаем ошибку - Bad Request?
			h.logger.Err(err).
				Str("func", "*Handler.WithHashing").
				Str("hash from header", hashFromHeader).
				Str("hashed body", hex.EncodeToString(hashedBody)).
				Msg("hash from header and hash from body are not equal")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		h.logger.Debug().
			Str("func", "*Handler.WithHashing").
			Str("hash from header", hashFromHeader).
			Str("hashed body", hex.EncodeToString(hashedBody)).
			Msg("hash from header and hash from body ARE EQUAL")

		// если хэши совпали - продвигаем дальше запрос
		next.ServeHTTP(w, r)

		// 2. Формируем ответ

		// 1. Есть ли ключ установленный при запуске?
		// =====Если нет =======
		// 		ничего не делаем
		//  ==== Если есть ключ =====
		//  Проверяем
	})
}

func (h *Handler) WriteResponseWithHashing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.logger.Debug().
			Str("func", "*Handler.WriteResponseWithHashing").
			Msg("WriteResponseWithHashing was called")

		// как поймать запрос ответный?
		// response writer !!! Он создан для ЗАПИСИ сервером ОТВЕТА НА ЗАПРОС
		if h.hasher == nil {
			h.logger.Debug().
				Str("func", "*Handler.WriteResponseWithHashing").
				Msg("skipping hashing of the response")
			next.ServeHTTP(w, r)
			return
		}

		// При наличии ключа на этапе формирования ответа сервер должен вычислять хеш и передавать его в HTTP-заголовке ответа с именем
		rw := &HashingResponseWriter{
			ResponseWriter: w,
			buf:            bytes.NewBuffer(nil),
			hasher:         h.hasher,
			logger:         h.logger,
			statusCode:     http.StatusOK, // default value if WriteHeader wasn't called
		}

		h.logger.Debug().
			Str("func", "*Handler.WriteResponseWithHashing").
			Any("response writer", rw).
			Msg("adding new response writer with hashing")

		next.ServeHTTP(rw, r)
	})
}

type HashingResponseWriter struct {
	http.ResponseWriter
	hasher            *utils.Hasher
	logger            *zerolog.Logger
	buf               *bytes.Buffer
	statusCode        int
	contentSize       int
	writeHeaderCalled bool
}

// WriteHeader что делать? Мне нужно записать только код ответа, но нужно заголовок при вызове write выполнить
// Значит не надо выполнять http.ResponseWriter - WriteHeader.
// Надо записать код ответа в структуру и ничего не делать больше
func (rw *HashingResponseWriter) WriteHeader(statusCode int) {
	rw.logger.Debug().
		Str("func", "*HashingResponseWriter.WriteHeader").
		Int("status code recived", statusCode).
		Msg("WriteHeader was called")
	rw.statusCode = statusCode
}

// Write записываем хэш в хэдер и пробрасываем дальше
func (rw *HashingResponseWriter) Write(response []byte) (int, error) {
	rw.logger.Debug().
		Str("func", "*HashingResponseWriter.Write").
		Int("status code", rw.statusCode).
		Msg("Write() was called")

	// 1. проверить не пустой ли массив байтов
	if response == nil {
		// ничего не делаем - отправляем тупо ответ
		rw.Header().Set("HashSHA256", "")
		rw.ResponseWriter.WriteHeader(rw.statusCode)
		rw.logger.Debug().
			Str("func", "*HashingResponseWriter.Write").
			Int("status code", rw.statusCode).
			Any("headers to send", rw.ResponseWriter.Header()).
			Msg("empty write was called")

		return rw.ResponseWriter.Write([]byte("empty body"))
	}

	// если не пустой - вычисляем хэш
	responseHash, err := rw.hasher.HashByteSlice(response)
	if err != nil {
		// возвращаем ошибку - Internal Server Error
		rw.logger.Err(err).
			Str("func", "*HashingResponseWriter.Write").
			Msg("error occurred during hashing")
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return rw.contentSize, err
	}

	rw.Header().Set("HashSHA256", hex.EncodeToString(responseHash))
	rw.WriteHeader(rw.statusCode)

	rw.logger.Debug().
		Str("func", "*HashingResponseWriter.Write").
		Int("status code", rw.statusCode).
		Any("headers to send", rw.ResponseWriter.Header()).
		Bytes("response", response).
		Msg("hashing completed")

	return rw.buf.Write(response)
}

// captureWriter нужен, чтобы перехватить тело ответа
//type captureWriter struct {
//	http.ResponseWriter
//	buf               bytes.Buffer
//	statusCode        int
//	logger            *zerolog.Logger
//	writeHeaderCalled bool
//}
//
//func (cw *captureWriter) Write(p []byte) (int, error) {
//	cw.buf.Write(p)
//	return cw.ResponseWriter.Write(p)
//}
//
//func (cw *captureWriter) WriteHeader(statusCode int) {
//	cw.writeHeaderCalled = true
//	cw.Header().Add("HashSHA256", "EXAMPLE")
//	cw.logger.Debug().
//		Str("func", "responseWriterWithHash.WriteHeader()").
//		Int("statusCode", statusCode).
//		Any("header", cw.Header()).Msg("")
//	cw.statusCode = statusCode
//	// Как сделать так, чтобы статус устанавливался даже в случае, если нет вызова метода Write()
//}

//// WithHashing проверяет входящий HashSHA256 и подписывает исходящий
//func (h *Handler) WithHashing(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		var body []byte
//		// === Проверка входящего хеша ===
//		if h.hasher != nil {
//			clientHash := r.Header.Get("HashSHA256")
//
//			if clientHash != "" { // если заголовок есть — проверяем
//				// читаем тело запроса
//				body, err := io.ReadAll(r.Body)
//				if err != nil {
//					http.Error(w, "failed to read body", http.StatusBadRequest)
//					return
//				}
//				_ = r.Body.Close()
//				r.Body = io.NopCloser(bytes.NewBuffer(body)) // восстанавливаем для хендлера
//
//				// вычисляем хеш
//				expectedBytes, err := h.hasher.HashByteSlice(body)
//				if err != nil {
//					http.Error(w, "failed to compute hash", http.StatusInternalServerError)
//					return
//				}
//				expected := hex.EncodeToString(expectedBytes)
//
//				// сравнение
//				if clientHash != expected {
//					http.Error(w, "invalid hash", http.StatusBadRequest)
//					return
//				}
//			}
//		}
//
//		// === Формирование ответа ===
//		if h.hasher == nil {
//			// если ключа нет — хеширование отключено
//			next.ServeHTTP(w, r)
//			return
//		}
//
//		// подменяем writer
//		cw := &captureWriter{ResponseWriter: w, logger: h.logger}
//		if body != nil {
//			next.ServeHTTP(cw, r)
//			// считаем хеш ответа
//			respHashBytes, err := h.hasher.HashByteSlice(cw.buf.Bytes())
//			if err != nil {
//				http.Error(w, "failed to compute response hash", http.StatusInternalServerError)
//				return
//			}
//			w.Header().Set("HashSHA256", hex.EncodeToString(respHashBytes))
//		} else {
//			next.ServeHTTP(w, r)
//		}
//	})
//}

//
//func (h *Handler) WithHashing(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
//		// if we don't have hashkey to check hash or do not have an http-body - do nothing
//		if h.hasher == nil || req.Body == nil {
//			next.ServeHTTP(w, req)
//			return
//		}
//		h.logger.Debug().Str("func", "*Handler.WithHashing").Msg("validating metric hash")
//
//		// check if hash exists in header
//		hashFromHeader := req.Header.Get("HashSHA256")
//
//		if hashFromHeader == "" {
//			h.logger.Error().Str("func", "*Handler.WithHashing").Any("headers", req.Header).Msg("empty hash header")
//			//http.Error(w, "empty hash", http.StatusBadRequest)
//			//return
//			//next.ServeHTTP(w, req)
//		} else {
//			// read request body
//			body, err := io.ReadAll(req.Body)
//			if err != nil {
//				h.logger.Err(err).Str("func", "*Handler.WithHashing").Msg("error during reading a body")
//				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
//				return
//			}
//			h.logger.Debug().Str("func", "*Handler.WithHashing").Bytes("http body", body).Msg("")
//
//			// return body to the request
//			req.Body = io.NopCloser(bytes.NewBuffer(body))
//
//			// get hash body
//			hashSliceFromBody, err := h.hasher.HashByteSlice(body)
//			if err != nil {
//				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
//				return
//			}
//			h.logger.Debug().Str("func", "*Handler.WithHashing").Str("hash from header", hashFromHeader).Str("hash from body", hex.EncodeToString(hashSliceFromBody)).Msg("")
//
//			// compare hash from header with calculated hash from body
//			if hashFromHeader != hex.EncodeToString(hashSliceFromBody) {
//				h.logger.Debug().Str("func", "*Handler.WithHashing").Msg("hash is not valid")
//				http.Error(w, "hashes do not match", http.StatusBadRequest)
//				return
//			}
//
//			h.logger.Debug().Str("func", "*Handler.WithHashing").Msg("metric hash is valid")
//
//		}
//		// catching response from server
//		rw := &responseWriterWithHash{ResponseWriter: w, hasher: h.hasher, logger: h.logger}
//
//		// call another middleware or handler
//		next.ServeHTTP(rw, req)
//
//		// если статус так и не был установлен — принудительно выставляем 200
//		if !rw.writeHeaderCalled {
//			rw.WriteHeader(http.StatusOK)
//		}
//	})
//}
//
//// responseWriterWithHash catches response from server
//type responseWriterWithHash struct {
//	http.ResponseWriter
//	hasher            *utils.Hasher
//	logger            *zerolog.Logger
//	statusCode        int
//	writeHeaderCalled bool
//}
//
//func (rw *responseWriterWithHash) WriteHeader(statusCode int) {
//	rw.writeHeaderCalled = true
//	rw.logger.Debug().
//		Str("func", "responseWriterWithHash.WriteHeader()").
//		Int("statusCode", statusCode).
//		Any("header", rw.Header()).Msg("")
//	rw.statusCode = statusCode
//	// Как сделать так, чтобы статус устанавливался даже в случае, если нет вызова метода Write()
//}
//
//func (rw *responseWriterWithHash) Write(p []byte) (int, error) {
//	rw.logger.Debug().
//		Str("func", "responseWriterWithHash.Write()").
//		Bytes("writer bytes", p).Msg("")
//
//	// hash body
//	if responseHash, err := rw.hasher.HashByteSlice(p); err == nil {
//		rw.Header().Set("HashSHA256", hex.EncodeToString(responseHash))
//	}
//
//	if rw.statusCode == 0 {
//		rw.statusCode = http.StatusOK
//	}
//	rw.ResponseWriter.WriteHeader(rw.statusCode)
//
//	rw.logger.Debug().
//		Str("func", "responseWriterWithHash.Write()").
//		Any("header", rw.Header()).
//		RawJSON("body", p).Msg("")
//
//	return rw.ResponseWriter.Write(p)
//}
