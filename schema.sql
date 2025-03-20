-- Tabla para almacenar los ESP32
CREATE TABLE IF NOT EXISTS esp32m (
    id INT PRIMARY KEY AUTO_INCREMENT,
    serial VARCHAR(50) NOT NULL UNIQUE
);

-- Tabla para almacenar los chats de Telegram asociados a ESP32
CREATE TABLE IF NOT EXISTS telegram_chats (
    id INT PRIMARY KEY AUTO_INCREMENT,
    chat_id BIGINT NOT NULL,
    esp32_serial VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (esp32_serial) REFERENCES esp32m(serial),
    UNIQUE KEY unique_chat_esp32 (chat_id, esp32_serial)
);

-- Tabla para almacenar las lecturas del sensor KY-026
CREATE TABLE IF NOT EXISTS ky_026 (
    idKY_026 INT PRIMARY KEY AUTO_INCREMENT,
    esp32_serial VARCHAR(50) NOT NULL,
    fecha_activacion VARCHAR(45) NOT NULL,
    estado VARCHAR(50) NOT NULL,
    FOREIGN KEY (esp32_serial) REFERENCES esp32m(serial)
);
