#!/bin/bash

# Test script for Employee Service API
# This script demonstrates how to test the multi-tenant employee service

set -e

# Configuration
BASE_URL="http://localhost:8000"
JWT_SECRET="${JWT_SECRET:-test-secret-key}"

echo "=== Employee Service API Test Script ==="
echo "Base URL: $BASE_URL"
echo "JWT Secret: $JWT_SECRET"
echo ""

# Function to generate JWT token
generate_jwt() {
    local user_id=$1
    local tenant_id=$2
    
    # Use the generate-jwt.go script
    go run scripts/generate-jwt.go "$JWT_SECRET" "$user_id" "$tenant_id" | grep -A 1 "Generated JWT Token:" | tail -1
}

echo "Generating JWT tokens for two different tenants..."
TOKEN_TENANT_A=$(generate_jwt "user-12" "tenant-a")
TOKEN_TENANT_B=$(generate_jwt "user-223" "tenant-b")

echo "Token for Tenant A (user-1): ${TOKEN_TENANT_A:0:50}..."
echo "Token for Tenant B (user-2): ${TOKEN_TENANT_B:0:50}..."
echo ""

# Test 1: Create employee in Tenant A
echo "=== Test 1: Create employee in Tenant A ==="
# Use a unique email with timestamp to avoid conflicts
TIMESTAMP=$(date +%s)
TEST_EMAIL_A="alice-test-${TIMESTAMP}@example.com"
EMPLOYEE_A=$(curl -s -X POST "$BASE_URL/api/v1/employees" \
  -H "Authorization: Bearer $TOKEN_TENANT_A" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL_A\",
    \"first_name\": \"Alice\",
    \"last_name\": \"Smith\"
  }")

echo "Created employee in Tenant A: $EMPLOYEE_A"
EMPLOYEE_A_ID=$(echo "$EMPLOYEE_A" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
if [ -z "$EMPLOYEE_A_ID" ]; then
    EMPLOYEE_A_ID=$(echo "$EMPLOYEE_A" | grep -o '"id":[0-9]*' | cut -d':' -f2)
fi
echo "Employee A ID: $EMPLOYEE_A_ID"
echo ""

# Test 2: Create employee in Tenant B
echo "=== Test 2: Create employee in Tenant B ==="
TEST_EMAIL_B="bob-test-${TIMESTAMP}@example.com"
EMPLOYEE_B=$(curl -s -X POST "$BASE_URL/api/v1/employees" \
  -H "Authorization: Bearer $TOKEN_TENANT_B" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL_B\",
    \"first_name\": \"Bob\",
    \"last_name\": \"Jones\"
  }")

echo "Created employee in Tenant B: $EMPLOYEE_B"
EMPLOYEE_B_ID=$(echo "$EMPLOYEE_B" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
if [ -z "$EMPLOYEE_B_ID" ]; then
    EMPLOYEE_B_ID=$(echo "$EMPLOYEE_B" | grep -o '"id":[0-9]*' | cut -d':' -f2)
fi
echo "Created employee in Tenant B: $EMPLOYEE_B"
echo "Employee B ID: $EMPLOYEE_B_ID"
echo ""

# Test 3: Try to access Tenant A employee with Tenant B token (should fail)
echo "=== Test 3: Multi-tenant Isolation Test ==="
echo "Attempting to access Tenant A employee ($EMPLOYEE_A_ID) with Tenant B token..."
RESULT=$(curl -s -w "\nHTTP_CODE:%{http_code}" "$BASE_URL/api/v1/employees/$EMPLOYEE_A_ID" \
  -H "Authorization: Bearer $TOKEN_TENANT_B")

HTTP_CODE=$(echo "$RESULT" | grep HTTP_CODE | cut -d: -f2)
if [ "$HTTP_CODE" = "404" ]; then
    echo "✓ PASS: Tenant isolation working (got 404 as expected)"
else
    echo "✗ FAIL: Expected 404, got $HTTP_CODE"
fi
echo ""

# Test 4: List employees in Tenant A
echo "=== Test 4: List employees in Tenant A ==="
curl -s "$BASE_URL/api/v1/employees?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN_TENANT_A" | jq .
echo ""

# Test 5: List employees in Tenant B
echo "=== Test 5: List employees in Tenant B ==="
curl -s "$BASE_URL/api/v1/employees?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN_TENANT_B" | jq .
echo ""

# Test 6: Create duplicate email in same tenant (should fail)
echo "=== Test 6: Duplicate email test in same tenant ==="
DUPLICATE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE_URL/api/v1/employees" \
  -H "Authorization: Bearer $TOKEN_TENANT_A" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL_A\",
    \"first_name\": \"Alice2\",
    \"last_name\": \"Smith2\"
  }")

HTTP_CODE=$(echo "$DUPLICATE" | grep HTTP_CODE | cut -d: -f2)
if [ "$HTTP_CODE" = "400" ]; then
    echo "✓ PASS: Duplicate email rejected (got 400 as expected)"
else
    echo "✗ FAIL: Expected 400, got $HTTP_CODE"
fi
echo ""

# Test 7: Create same email in different tenant (should succeed)
echo "=== Test 7: Same email in different tenant test ==="
# Try to create the same email from Tenant A in Tenant B - this should work!
SAME_EMAIL=$(curl -s -X POST "$BASE_URL/api/v1/employees" \
  -H "Authorization: Bearer $TOKEN_TENANT_B" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL_A\",
    \"first_name\": \"Alice\",
    \"last_name\": \"Brown-TenantB\"
  }")

echo "Response: $SAME_EMAIL"
SAME_EMAIL_ID=$(echo "$SAME_EMAIL" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
if [ -z "$SAME_EMAIL_ID" ]; then
    SAME_EMAIL_ID=$(echo "$SAME_EMAIL" | grep -o '"id":[0-9]*' | cut -d':' -f2)
fi

if [ -n "$SAME_EMAIL_ID" ]; then
    echo "✓ PASS: Same email ($TEST_EMAIL_A) created in different tenant (ID: $SAME_EMAIL_ID)"
    echo "  This proves email uniqueness is scoped per tenant!"
else
    echo "✗ FAIL: Could not create same email in different tenant"
    echo "  Response: $SAME_EMAIL"
fi
echo ""

# Test 8: Update employee
echo "=== Test 8: Update employee ==="
if [ -n "$EMPLOYEE_A_ID" ]; then
    UPDATED_EMAIL="alice-updated-${TIMESTAMP}@example.com"
    curl -s -X PUT "$BASE_URL/api/v1/employees/$EMPLOYEE_A_ID" \
      -H "Authorization: Bearer $TOKEN_TENANT_A" \
      -H "Content-Type: application/json" \
      -d "{
        \"id\": \"$EMPLOYEE_A_ID\",
        \"email\": \"$UPDATED_EMAIL\",
        \"first_name\": \"Alice\",
        \"last_name\": \"Smith-Updated\"
      }" | jq .
    echo ""
    
    # Test 9: Get employee by email
    echo "=== Test 9: Get employee by email (using list endpoint) ==="
    curl -s "$BASE_URL/api/v1/employees?email=$UPDATED_EMAIL" \
      -H "Authorization: Bearer $TOKEN_TENANT_A" | jq .
else
    echo "Skipping Tests 8-9: No employee ID from Test 1"
fi
echo ""

# Test 10: Invalid UUID format
echo "=== Test 10: Invalid UUID Format ==="
INVALID_UUID=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
  "$BASE_URL/api/v1/employees/not-a-uuid" \
  -H "Authorization: Bearer $TOKEN_TENANT_A")

HTTP_CODE=$(echo "$INVALID_UUID" | grep HTTP_CODE | cut -d: -f2)
if [ "$HTTP_CODE" = "400" ]; then
    echo "✓ PASS: Invalid UUID rejected (got 400 as expected)"
else
    echo "✗ FAIL: Expected 400, got $HTTP_CODE"
    echo "Response: $(echo "$INVALID_UUID" | grep -v HTTP_CODE)"
fi
echo ""

# Test 11: Validate UUID format from created employee
echo "=== Test 11: UUID Format Validation ==="
if [[ -n "$EMPLOYEE_A_ID" && "$EMPLOYEE_A_ID" =~ ^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$ ]]; then
    echo "✓ PASS: Employee ID is a valid UUID v4 format: $EMPLOYEE_A_ID"
else
    echo "✗ FAIL: Employee ID is not a valid UUID format: $EMPLOYEE_A_ID"
fi
echo ""

# ========================================
# VALIDATION TESTS
# ========================================

echo "========================================="
echo "         VALIDATION TESTS"
echo "========================================="
echo ""

# Test V1: Invalid email format
echo "=== Test V1: Invalid Email Format ==="
INVALID_EMAIL=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE_URL/api/v1/employees" \
  -H "Authorization: Bearer $TOKEN_TENANT_A" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "not-an-email",
    "first_name": "Test",
    "last_name": "User"
  }')

HTTP_CODE=$(echo "$INVALID_EMAIL" | grep HTTP_CODE | cut -d: -f2)
if [ "$HTTP_CODE" = "400" ]; then
    echo "✓ PASS: Invalid email rejected (got 400)"
else
    echo "✗ FAIL: Expected 400, got $HTTP_CODE"
fi
echo "Response: $(echo "$INVALID_EMAIL" | grep -v HTTP_CODE | head -1)"
echo ""

# Test V2: Empty required fields
echo "=== Test V2: Empty Name Fields ==="
EMPTY_NAME=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE_URL/api/v1/employees" \
  -H "Authorization: Bearer $TOKEN_TENANT_A" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "first_name": "",
    "last_name": ""
  }')

HTTP_CODE=$(echo "$EMPTY_NAME" | grep HTTP_CODE | cut -d: -f2)
if [ "$HTTP_CODE" = "400" ]; then
    echo "✓ PASS: Empty names rejected (got 400)"
else
    echo "✗ FAIL: Expected 400, got $HTTP_CODE"
fi
echo "Response: $(echo "$EMPTY_NAME" | grep -v HTTP_CODE | head -1)"
echo ""

# Test V3: Email too short
echo "=== Test V3: Email Too Short ==="
SHORT_EMAIL=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE_URL/api/v1/employees" \
  -H "Authorization: Bearer $TOKEN_TENANT_A" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "a@",
    "first_name": "Test",
    "last_name": "User"
  }')

HTTP_CODE=$(echo "$SHORT_EMAIL" | grep HTTP_CODE | cut -d: -f2)
if [ "$HTTP_CODE" = "400" ]; then
    echo "✓ PASS: Short email rejected (got 400)"
else
    echo "✗ FAIL: Expected 400, got $HTTP_CODE"
fi
echo "Response: $(echo "$SHORT_EMAIL" | grep -v HTTP_CODE | head -1)"
echo ""

# Test V4: Name with invalid characters
echo "=== Test V4: Invalid Name Characters ==="
INVALID_NAME=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE_URL/api/v1/employees" \
  -H "Authorization: Bearer $TOKEN_TENANT_A" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "first_name": "Test123",
    "last_name": "User@#$"
  }')

HTTP_CODE=$(echo "$INVALID_NAME" | grep HTTP_CODE | cut -d: -f2)
if [ "$HTTP_CODE" = "400" ]; then
    echo "✓ PASS: Invalid name characters rejected (got 400)"
else
    echo "✗ FAIL: Expected 400, got $HTTP_CODE"
fi
echo "Response: $(echo "$INVALID_NAME" | grep -v HTTP_CODE | head -1)"
echo ""

# Test V5: Name too long
echo "=== Test V5: Name Too Long ==="
LONG_NAME=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE_URL/api/v1/employees" \
  -H "Authorization: Bearer $TOKEN_TENANT_A" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"test@example.com\",
    \"first_name\": \"$(printf 'A%.0s' {1..101})\",
    \"last_name\": \"User\"
  }")

HTTP_CODE=$(echo "$LONG_NAME" | grep HTTP_CODE | cut -d: -f2)
if [ "$HTTP_CODE" = "400" ]; then
    echo "✓ PASS: Long name rejected (got 400)"
else
    echo "✗ FAIL: Expected 400, got $HTTP_CODE"
fi
echo ""

# Test V6: Invalid pagination (page = 0)
echo "=== Test V6: Pagination Default (page=0) ==="
PAGE_ZERO=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
  "$BASE_URL/api/v1/employees?page=0&page_size=20" \
  -H "Authorization: Bearer $TOKEN_TENANT_A")

HTTP_CODE=$(echo "$PAGE_ZERO" | grep HTTP_CODE | cut -d: -f2)
if [ "$HTTP_CODE" = "200" ]; then
    echo "✓ PASS: page=0 accepted and defaults to page 1 (got 200)"
    echo "  This is good UX - clients don't need to explicitly set page=1"
else
    echo "✗ FAIL: Expected 200, got $HTTP_CODE"
fi
echo ""

# Test V7: Invalid pagination (page_size too large)
echo "=== Test V7: Invalid Pagination (page_size=1000) ==="
INVALID_SIZE=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
  "$BASE_URL/api/v1/employees?page=1&page_size=1000" \
  -H "Authorization: Bearer $TOKEN_TENANT_A")

HTTP_CODE=$(echo "$INVALID_SIZE" | grep HTTP_CODE | cut -d: -f2)
if [ "$HTTP_CODE" = "400" ]; then
    echo "✓ PASS: Invalid page_size rejected (got 400)"
else
    echo "✗ FAIL: Expected 400, got $HTTP_CODE"
fi
echo ""

# Test V8: Same email in merge (business validation)
echo "=== Test V8: Merge Same Email (Business Validation) ==="
SAME_EMAIL_MERGE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE_URL/api/v1/employees/merge" \
  -H "Authorization: Bearer $TOKEN_TENANT_A" \
  -H "Content-Type: application/json" \
  -d '{
    "primary_email": "test@example.com",
    "secondary_email": "test@example.com"
  }')

HTTP_CODE=$(echo "$SAME_EMAIL_MERGE" | grep HTTP_CODE | cut -d: -f2)
RESPONSE=$(echo "$SAME_EMAIL_MERGE" | grep -v HTTP_CODE)
if [ "$HTTP_CODE" = "400" ]; then
    if echo "$RESPONSE" | grep -q "INVALID_MERGE"; then
        echo "✓ PASS: Business validation rejected same email merge (got 400 with INVALID_MERGE)"
    else
        echo "✓ PASS: Same email merge rejected (got 400)"
    fi
else
    echo "✗ FAIL: Expected 400, got $HTTP_CODE"
fi
echo "Response: $(echo "$RESPONSE" | head -1)"
echo ""

# Test V9: Valid name with special characters (should pass)
echo "=== Test V9: Valid Name with Hyphens and Apostrophes ==="
VALID_NAME=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE_URL/api/v1/employees" \
  -H "Authorization: Bearer $TOKEN_TENANT_A" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"mary-jane-${TIMESTAMP}@example.com\",
    \"first_name\": \"Mary-Jane\",
    \"last_name\": \"O'Brien-Smith\"
  }")

HTTP_CODE=$(echo "$VALID_NAME" | grep HTTP_CODE | cut -d: -f2)
if [ "$HTTP_CODE" = "200" ]; then
    echo "✓ PASS: Valid special characters accepted (got 200)"
else
    echo "✗ FAIL: Expected 200, got $HTTP_CODE"
fi
echo ""

echo "========================================="
echo "   VALIDATION TESTS COMPLETED"
echo "========================================="
echo ""

echo "=== All tests completed ==="

