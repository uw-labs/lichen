internal/license/db/licenses.db:
	go install github.com/google/licenseclassifier/tools/license_serializer
	license_serializer -output ./internal/license/db/
