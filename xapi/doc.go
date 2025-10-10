// Package xapi provides a type-safe lightweight HTTP API framework for Go.
//
// Most HTTP handlers follow the same pattern - decode JSON, extract headers/params,
// validate, call business logic, encode response. xapi codifies that pattern using
// generics, so you write less but get more type safety. Your request and response
// types define the API contract. The optional interfaces provide flexibility when needed.
//
// The result: handlers that are mostly business logic, with HTTP operations abstracted
// away into a lightweight framework. You can use it with your existing HTTP router and
// server, keeping all existing middlewares and error handling.
//
// # Core Types
//
// [Endpoint] is the main type that wraps your [EndpointHandler] and applies middleware
// and error handling. Create endpoints using [NewEndpoint] with your handler and optional
// configuration via [EndpointOption] values.
//
// [EndpointFunc] is a function type that implements [EndpointHandler], providing a
// convenient way to create handlers from functions.
//
// # Optional Interfaces
//
// xapi defines four optional interfaces. Implement them on request and response types
// only when needed:
//
// [Validator] runs after JSON decoding to validate the request. You can use any validation
// library here.
//
// [Extracter] pulls data from the HTTP request that isn't in the JSON body, such as headers,
// route path params, or query strings.
//
// [StatusSetter] controls the HTTP status code. The default is 200, but you can override it
// to return 201 for creation, 204 for no content, etc.
//
// [RawWriter] bypasses JSON encoding entirely for HTML, or binary responses.
// Use this when you need full control over the response format.
//
// # Middleware
//
// Middleware works exactly like standard http.Handler middleware. Any middleware you're
// already using will work. Stack them in the order you need using [WithMiddleware]. They
// wrap the endpoint cleanly, keeping auth, logging, and metrics separate from your
// business logic. Use [MiddlewareFunc] to convert functions to middleware, or implement
// [MiddlewareHandler] for custom middleware types.
//
// # Error Handling
//
// Default behavior is a 500 status with the error text. Customize this using
// [WithErrorHandler] to distinguish validation errors from auth failures, map them to
// appropriate status codes, and format them consistently. Implement the [ErrorHandler]
// interface or use [ErrorFunc] for simple function-based handlers. The default error
// handling is provided by [DefaultErrorHandler].
package xapi
