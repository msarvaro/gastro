#!/bin/bash

# Database connection parameters - update these as needed
DB_USER="postgres"
DB_NAME="user_management"
DB_HOST="localhost"
DB_PORT="5432"

echo "Applying migrations to database $DB_NAME..."

# Apply the migrations
for file in $(ls *.sql | sort); do
    echo "Applying migration: $file"
    psql -U $DB_USER -h $DB_HOST -p $DB_PORT -d $DB_NAME -f $file
    if [ $? -ne 0 ]; then
        echo "Migration failed: $file"
        exit 1
    fi
done

echo "All migrations applied successfully!" 