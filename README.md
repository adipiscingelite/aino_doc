# backend
Base URL = http://localhost:1234

# list all form
// You can change it according to your data stored in document_ms.
- Dampak Analisa -> filtered by document_code = 'DA' // You can change it in service.go.
- ITCM -> filtered by document_code = 'ITCM'
- Berita Acara -> filtered by document_code = 'BA'

# auth
// You can change it in middleware, according to your data stored in role_ms.
Role required:
- Member -> middleware with role_code = 'M' 
- Admin -> middleware with role_code = 'A'
- Superadmin -> middleware with role_code = 'SA'