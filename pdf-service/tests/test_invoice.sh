curl -X POST http://localhost:8080/generate-pdf \
  -H "Content-Type: application/json" \
  -d '{
  "template": "form",
  "data": {
    "CompanyName": "ToCode Pvt Ltd",
    "CompanyAddress": "Chennai, India",

    "CustomerName": "Ezhil",
    "CustomerAddress": "Ambur, Tamil Nadu",

    "InvoiceNumber": "INV-001",
    "Date": "2026-04-09",
    "DueDate": "2026-04-15",

    "Items": [
      {
        "Name": "Product A",
        "Quantity": 2,
        "Price": 500,
        "Total": 1000
      },
      {
        "Name": "Product B",
        "Quantity": 1,
        "Price": 300,
        "Total": 300
      }
    ],

    "Subtotal": 1300,
    "Tax": 130,
    "Total": 1430,

    "FooterNote": "Thank you for your business!"
  }
}' --output invoice.pdf





curl -X POST http://localhost:8080/generate-pdf \
  -H "Content-Type: application/json" \
  -d '{
  "template": "ezhil",
  "data": {
    
  }
}' --output ezhil.pdf


curl -X POST http://localhost:8080/generate-pdf \
  -H "Content-Type: application/json" \
  -d '{
    "template": "ezhil",
    "pages": ["terms", "privacy", "signature"],
    "data": {
      "CompanyName": "ToCode",
      "CustomerName": "Ezhil"
    }
  }' --output invoice.pdf