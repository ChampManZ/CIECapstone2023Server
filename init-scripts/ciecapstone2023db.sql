CREATE DATABASE IF NOT EXISTS ciecapstone2023db;
USE ciecapstone2023db;

CREATE TABLE Announcer (
    AnnouncerID INT PRIMARY KEY AUTO_INCREMENT,
    AnnouncerName VARCHAR(255),
    AnnouncerScript VARCHAR(255),
    SessionOfAnnounce VARCHAR(255),
    FirstOrder INT,
    LastOrder INT,
    IsBreak Boolean DEFAULT false
);

CREATE TABLE Certificate (
    CertificateID INT PRIMARY KEY AUTO_INCREMENT,
    Degree VARCHAR(255),
    Faculty VARCHAR(255),
    Major VARCHAR(255),
    Honor VARCHAR(255)
);

CREATE TABLE Student (
    StudentID INT PRIMARY KEY,
    OrderOfReceive INT,
    Firstname VARCHAR(255),
    Surname VARCHAR(255),
    NamePronunciation VARCHAR(255),
    CertificateID INT,
    AnnouncerID INT,
    FOREIGN KEY (AnnouncerID) REFERENCES Announcer(AnnouncerID),
    FOREIGN KEY (CertificateID) REFERENCES Certificate(CertificateID)
);

CREATE TABLE NameRead (
    NameReadStudentID INT PRIMARY KEY,
    SavedNameRead VARCHAR(255)
);

CREATE TABLE Counter (
    ID INT PRIMARY KEY,
    CurrentValue INT
);