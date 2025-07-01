package client

import "errors"

var ErrClientNotFound = errors.New("client not found")
var ErrClientAlreadyExists = errors.New("client already exists")
var ErrClientNameIsRequired = errors.New("client name is required")
var ErrClientWhatsAppNumberIsRequired = errors.New("client whatsapp number is required")
var ErrClientCreatedAtIsRequired = errors.New("client created at is required")
var ErrClientUpdatedAtIsRequired = errors.New("client updated at is required")
var ErrClientIDIsRequired = errors.New("client id is required")
var ErrFailedToGetClients = errors.New("failed to get clients")
