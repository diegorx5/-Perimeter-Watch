# -Perimeter-Watch
Sistema de vigilancia perimetral con cámara térmica FLIR y tracking PTZ automático, escrito en Go. Detecta firmas de calor humano en tiempo real y mueve la cámara para centrar el objetivo — sin intervención manual

¿Qué hace?

Captura frames térmicos de la cámara FLIR via RTSP
Analiza cada frame buscando regiones de alta temperatura
Calcula el offset del objetivo respecto al centro del frame
Mueve la cámara PTZ proporcionalmente para hacer tracking
Muestra todo en un dashboard web en tiempo real

Requisitos

Go 1.21+
Cámara FLIR con soporte RTSP y ONVIF (probado con FLIR en IP 192.168.40.84)
Red local con acceso a la cámara


Instalación
bashgit clone https://github.com/TU_USUARIO/perimeter-watch.git
cd perimeter-watch
go mod download

Configuración
Edita las constantes al inicio de main.go:
goconst (
    CAMERA_IP   = "192.168.40.84"  // IP de tu cámara
    CAMERA_USER = "admin"           // Usuario
    CAMERA_PASS = "admin"           // Contraseña
)

Uso
bashgo run main.go
El sistema arranca, mueve la cámara a posición home y entra en modo patrulla continua.
Cuando detecta una firma térmica imprime en consola:
ALERTA - FIRMA TERMICA DETECTADA
Objetivo: centro(310,220) temp=36.5C
PTZ X:0.08 Y:0.04
El dashboard web se abre automáticamente en http://localhost:8080.

Estructura del proyecto
├── main.go              # Loop principal de patrulla
├── camera/              # Captura RTSP y control PTZ ONVIF
├── detector/            # Detección y dibujo de firmas térmicas
├── dashboard/           # Dashboard web en tiempo real
├── go.mod
└── go.sum

Stack

Go — backend principal
ONVIF (use-go/onvif) — control PTZ
RTSP — captura de video térmico
Procesamiento de imagen — detección de blobs por temperatura
