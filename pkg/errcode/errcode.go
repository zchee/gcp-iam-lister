// Package errcode handles the gRPC or google.golang.org/api packages error codes.
package errcode

import (
	"fmt"
	"net/http"

	"google.golang.org/api/googleapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func FromError(err error) error {
	switch code := status.Code(err); code {
	case codes.NotFound:
		return status.Error(code, http.StatusText(http.StatusNotFound))

	case codes.PermissionDenied:
		return status.Error(code, http.StatusText(http.StatusForbidden))

	default:
		if apierr, ok := err.(*googleapi.Error); ok {
			switch code := apierr.Code; code {
			case http.StatusForbidden:
				return fmt.Errorf(http.StatusText(http.StatusForbidden))
			default:
				return fmt.Errorf(http.StatusText(code))
			}
		}
	}

	return status.Error(codes.Unknown, "unknown error")
}
