# Bitácora de Decisiones

| Hito/Fecha | Decisión | Alternativas | Justificación Técnica | Uso de IA |
| :--- | :--- | :--- | :--- | :--- |
| Hito 3 / 06-07-2026 | Definición de interfaz Operator y Row | Structs estáticos | Uso de slice de 'any' para soportar tipado dinámico y nulos (nil). Método Close() para evitar fugas al leer CSV. | Ninguno |