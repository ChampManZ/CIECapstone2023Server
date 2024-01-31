import mysql.connector
import json
import requests

def get_data():
    url = "http://pundit.reg.kmitl.ac.th/pundit/api/index.php?function=get-data-ceremony&year=2565"
    response = requests.get(url)
    if response.status_code == 200:
        data = response.json()
    return data

db_config = {
    'host': 'localhost',
    'user': 'root',
    'password': 'admin123',
    'database': 'ciecapstone2023db'
}

db = mysql.connector.connect(**db_config)
cursor = db.cursor()

json_data = get_data()

# Load data into Database
# for item in json_data:
#     student_id = int(item['student_id'])
#     first_name = item['name']
#     surname = item['surname']
#     name_pronunciation = item['name_read']
#     order_of_receive = int(item['receive_order'])
#     degree = item['cer_name']
#     faculty = item['faculty_name']
#     major = item['curr_name']
#     honor = item['honor']

#     cursor.execute("""
#         INSERT INTO Certificate (Degree, Faculty, Major, Honor)
#         VALUES (%s, %s, %s, %s)
#     """, (degree, faculty, major, honor))
#     certificate_id = cursor.lastrowid  # Get the last inserted id

#     cursor.execute("""
#         INSERT INTO Student (StudentID, OrderOfReceive, Firstname, Surname, 
#                              NamePronunciation, CertificateID)
#         VALUES (%s, %s, %s, %s, %s, %s)
#     """, (student_id, order_of_receive, first_name, surname, name_pronunciation, certificate_id))

#     db.commit()

## kill conn
# cursor.close()
# db.close()

# Query student entries
query = """
    SELECT 
        s.StudentID,
        s.OrderOfReceive,
        s.Firstname,
        s.Surname,
        s.NamePronunciation,
        c.Degree,
        c.Faculty,
        c.Major,
        c.Honor,
        a.AnnouncerName,
        a.AnnouncerPos,
        a.SessionOfAnnounce
    FROM 
        Student s
    JOIN 
        Certificate c ON s.CertificateID = c.CertificateID
    LEFT JOIN 
        Announcer a ON c.AnnouncerID = a.AnnouncerID
    LIMIT 100;
"""

# Get unique Faculty
# query = "SELECT DISTINCT Faculty FROM Certificate;"

cursor.execute(query)
students = cursor.fetchall()
for student in students:
    print(student)

cursor.close()
db.close()
