# Step 1: Run Container
docker-compose up -d

# Step 2: Copy Script to Container
docker cp ciecapstone2023db.sql sql-db-1:ciecapstone2023db.sql 

# Step 3: Access MySQL Shell
docker exec -it sql-db-1 mysql -uroot -p

# Step 4: Execute SQL Script
source /ciecapstone2023db.sql

# Step 5: Verify
SHOW TABLES;

# UPDATE 7/1/2025
All files in this folder are deprecated and should be marked for delete on future update, due to containerization and the use of init-scripts