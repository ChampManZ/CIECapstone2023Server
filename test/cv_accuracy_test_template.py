import cv2
import numpy as np
import datetime
import json

# Initialize the video capture
cap = cv2.VideoCapture(0)

# Define the color range for detection (blue in this example)
lower_blue = np.array([110, 50, 50])
upper_blue = np.array([130, 255, 255])

# Placeholder for simulation purposes (manual or conditional logic to identify correct detections)
# In a real scenario, this could be replaced with actual validation against a gold standard.
is_correct_detection = lambda area: area > 500  # Simulating correctness based on area for demonstration

# Tracking variables
total_detections = 0
correct_detections = 0
detections = []

while True:
    ret, frame = cap.read()
    if not ret:
        break

    hsv_frame = cv2.cvtColor(frame, cv2.COLOR_BGR2HSV)
    mask = cv2.inRange(hsv_frame, lower_blue, upper_blue)
    result = cv2.bitwise_and(frame, frame, mask=mask)

    # Detect contours to identify objects
    contours, _ = cv2.findContours(mask, cv2.RETR_TREE, cv2.CHAIN_APPROX_SIMPLE)
    for contour in contours:
        area = cv2.contourArea(contour)
        total_detections += 1
        if is_correct_detection(area):
            correct_detections += 1
            x, y, w, h = cv2.boundingRect(contour)
            cv2.rectangle(frame, (x, y), (x+w, y+h), (0, 255, 0), 2)
            detections.append({"timestamp": str(datetime.datetime.now()), "coordinates": [x, y, w, h], "correct": True})
        else:
            detections.append({"timestamp": str(datetime.datetime.now()), "coordinates": [x, y, w, h], "correct": False})

    cv2.imshow('Frame', frame)
    if cv2.waitKey(1) & 0xFF == ord('q'):
        break

cap.release()
cv2.destroyAllWindows()

# Calculate correctness
correctness = correct_detections / total_detections if total_detections else 0

# Assuming a simple confusion matrix for demonstration purposes:
# True positives (TP): correct_detections
# False positives (FP): Assumed as detections that were not correct
# False negatives (FN) and True negatives (TN) are not directly measurable in this setup without additional context
confusion_matrix = {
    "TP": correct_detections,
    "FP": total_detections - correct_detections,
    "FN": 0,  # Placeholder, as we cannot measure without gold standard comparison
    "TN": 0   # Placeholder
}

# Log including confusion matrix and correctness
log_data = {
    "detections": detections,
    "correctness": correctness,
    "confusion_matrix": confusion_matrix
}

# Save to JSON file
with open("detection_log_with_evaluation.json", 'w') as file:
    json.dump(log_data, file, indent=4)