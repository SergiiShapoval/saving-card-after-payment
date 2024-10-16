curl -X POST http://localhost:4242/capture-payment-intent \
  -H "Content-Type: application/json" \
  -d '{
    "paymentIntentID": "pi_3Q9hpOAJlbf9cOtY3OQoDHb5",
    "amount": 100
  }'