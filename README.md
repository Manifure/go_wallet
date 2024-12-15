# go_wallet

Пример работы
POST /api/v1/wallet

Запрос:

POST /api/v1/wallet
Content-Type: application/json

{
  "walletId": "a3e1c8ac-4b6e-4f9e-9a02-2bdae5f2e0ee",
  "operationType": "DEPOSIT",
  "amount": 1000
}

Ответ:

HTTP/1.1 200 OK
Операция успешно выполнена. Новый баланс: 1000.00
GET /api/v1/wallets/{walletId}

Запрос:

GET /api/v1/wallets/a3e1c8ac-4b6e-4f9e-9a02-2bdae5f2e0ee

Ответ:

{
  "walletId": "a3e1c8ac-4b6e-4f9e-9a02-2bdae5f2e0ee",
  "balance": 1000.0
}
