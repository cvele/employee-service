#!/bin/bash

# Test script for NATS event publishing
set -e

BASE_URL="http://localhost:8000"
JWT_SECRET="${JWT_SECRET:-test-secret}"

echo "========================================="
echo "    NATS Event Publishing Test"
echo "========================================="
echo ""

# Check if NATS is running
echo "Checking NATS status..."
if ! docker ps | grep -q employee-service-nats; then
    echo "Starting NATS..."
    make nats-up
    echo "Waiting for NATS to be ready..."
    sleep 5
fi

# Check NATS health
if curl -s http://localhost:8222/healthz > /dev/null 2>&1; then
    echo "✓ NATS is running and healthy"
else
    echo "✗ NATS health check failed"
    exit 1
fi
echo ""

# Start consumer in background
echo "Starting event consumer..."
go run cmd/consumer/main.go > /tmp/nats-consumer.log 2>&1 &
CONSUMER_PID=$!

# Give consumer time to start
sleep 2

if ps -p $CONSUMER_PID > /dev/null; then
    echo "✓ Consumer started (PID: $CONSUMER_PID)"
else
    echo "✗ Consumer failed to start"
    exit 1
fi
echo ""

# Generate test JWT
echo "Generating test JWT..."
TOKEN=$(go run scripts/generate-jwt.go "$JWT_SECRET" "user-test" "tenant-test" 2>/dev/null | grep "^eyJ" | head -1)

if [ -z "$TOKEN" ]; then
    echo "✗ Failed to generate JWT token"
    kill $CONSUMER_PID 2>/dev/null || true
    exit 1
fi
echo "✓ JWT token generated"
echo ""

# Test 1: Create employee (should trigger employee.created event)
echo "========================================="
echo "Test 1: Create Employee Event"
echo "========================================="
TIMESTAMP=$(date +%s)
CREATE_RESULT=$(curl -s -X POST "$BASE_URL/api/v1/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"event-test-${TIMESTAMP}@example.com\",
    \"first_name\": \"Event\",
    \"last_name\": \"Test\"
  }")

EMPLOYEE_ID=$(echo "$CREATE_RESULT" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

if [ -n "$EMPLOYEE_ID" ]; then
    echo "✓ Employee created: $EMPLOYEE_ID"
else
    echo "✗ Failed to create employee"
    echo "Response: $CREATE_RESULT"
fi

sleep 2
echo ""

# Test 2: Update employee (should trigger employee.updated event)
if [ -n "$EMPLOYEE_ID" ]; then
    echo "========================================="
    echo "Test 2: Update Employee Event"
    echo "========================================="
    UPDATE_RESULT=$(curl -s -X PUT "$BASE_URL/api/v1/employees/$EMPLOYEE_ID" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"id\": \"$EMPLOYEE_ID\",
        \"first_name\": \"EventUpdated\",
        \"last_name\": \"TestUpdated\"
      }")
    
    if echo "$UPDATE_RESULT" | grep -q "EventUpdated"; then
        echo "✓ Employee updated"
    else
        echo "✗ Failed to update employee"
    fi
    
    sleep 2
    echo ""
fi

# Test 3: Merge employees (should trigger employee.merged event)
echo "========================================="
echo "Test 3: Merge Employees Event"
echo "========================================="

# Create second employee to merge
CREATE_SECOND=$(curl -s -X POST "$BASE_URL/api/v1/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"merge-test-${TIMESTAMP}@example.com\",
    \"first_name\": \"Merge\",
    \"last_name\": \"Test\"
  }")

sleep 1

# Merge them
MERGE_RESULT=$(curl -s -X POST "$BASE_URL/api/v1/employees/merge" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"primary_email\": \"event-test-${TIMESTAMP}@example.com\",
    \"secondary_email\": \"merge-test-${TIMESTAMP}@example.com\"
  }")

if echo "$MERGE_RESULT" | grep -q "secondaryEmails"; then
    echo "✓ Employees merged"
    # Extract and show the secondary emails
    SECONDARY_EMAILS=$(echo "$MERGE_RESULT" | grep -o '"secondaryEmails":\[[^]]*\]' | sed 's/"secondaryEmails":/Secondary emails: /')
    echo "  $SECONDARY_EMAILS"
else
    echo "✗ Failed to merge employees"
    echo "Response: $MERGE_RESULT"
fi

sleep 2
echo ""

# Test 4: Delete employee (should trigger employee.deleted event)
if [ -n "$EMPLOYEE_ID" ]; then
    echo "========================================="
    echo "Test 4: Delete Employee Event"
    echo "========================================="
    DELETE_RESULT=$(curl -s -X DELETE "$BASE_URL/api/v1/employees/$EMPLOYEE_ID" \
      -H "Authorization: Bearer $TOKEN")
    
    if echo "$DELETE_RESULT" | grep -q "success"; then
        echo "✓ Employee deleted"
    else
        echo "✗ Failed to delete employee"
    fi
    
    sleep 2
    echo ""
fi

# Stop consumer and show logs
echo "========================================="
echo "    Event Consumer Output"
echo "========================================="
kill $CONSUMER_PID 2>/dev/null || true
sleep 1

# Show consumer logs
if [ -f /tmp/nats-consumer.log ]; then
    cat /tmp/nats-consumer.log
    rm /tmp/nats-consumer.log
else
    echo "No consumer logs found"
fi

echo ""
echo "========================================="
echo "    Test Summary"
echo "========================================="
echo ""
echo "Expected events received by consumer:"
echo "  1. employee.created  - First employee"
echo "  2. employee.created  - Second employee (for merge)"
echo "  3. employee.updated  - Updated employee"
echo "  4. employee.merged   - Merged employees"
echo "  5. employee.deleted  - Deleted employee"
echo ""
echo "Check the output above for event details"
echo ""
echo "To manually test:"
echo "  Terminal 1: make consumer"
echo "  Terminal 2: Run API operations"
echo ""
echo "NATS monitoring: http://localhost:8222"
echo ""

