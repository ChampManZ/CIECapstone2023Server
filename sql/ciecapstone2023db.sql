USE ciecapstone2023db;

CREATE TABLE Announcer (
    AnnouncerID INT PRIMARY KEY AUTO_INCREMENT,
    AnnouncerName VARCHAR(255),
    AnnouncerPos VARCHAR(255),
    SessionOfAnnounce VARCHAR(255),
    ReceivingOrder INT
);

CREATE TABLE Certificate (
    CertificateID INT PRIMARY KEY AUTO_INCREMENT,
    Degree VARCHAR(255),
    Faculty VARCHAR(255),
    Major VARCHAR(255),
    Honor VARCHAR(255),
    AnnouncerID INT,
    FOREIGN KEY (AnnouncerID) REFERENCES Announcer(AnnouncerID)
);

CREATE TABLE Student (
    StudentID INT PRIMARY KEY,
    OrderOfReceive INT,
    CertificateOrder INT,
    Firstname VARCHAR(255),
    Surname VARCHAR(255),
    NamePronunciation VARCHAR(255),
    CertificateID INT,
    Notes TEXT,
    FOREIGN KEY (CertificateID) REFERENCES Certificate(CertificateID)
);

CREATE TABLE Counter (
    ID INT PRIMARY KEY,
    CurrentValue INT
);