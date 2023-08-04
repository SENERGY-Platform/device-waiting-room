/*
 * Copyright 2023 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mocks

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
)

func DeviceManager(ctx context.Context, wg *sync.WaitGroup, onCall func(path string, body []byte, err error) (resp []byte, code int)) (url string) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		resp, code := onCall(request.URL.Path, body, err)
		if code == http.StatusOK {
			if resp != nil {
				writer.Write(resp)
				return
			} else {
				writer.WriteHeader(code)
				return
			}
		}
		if resp != nil {
			http.Error(writer, string(resp), code)
			return
		} else {
			writer.WriteHeader(code)
			return
		}
	}))
	wg.Add(1)
	go func() {
		<-ctx.Done()
		server.Close()
		wg.Done()
	}()
	return server.URL
}
