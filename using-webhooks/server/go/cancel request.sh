curl -X POST http://localhost:4242/cancel-payment-intent \
  -H "Content-Type: application/json" \
  -d '{
    "paymentIntentID": "pi_3QA9K5AJlbf9cOtY16H5XcaW"
  }'