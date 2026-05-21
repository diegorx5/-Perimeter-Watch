package dashboard

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type AlertEvent struct {
	Timestamp   string  `json:"timestamp"`
	CenterX     int     `json:"centerX"`
	CenterY     int     `json:"centerY"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	Temperature float64 `json:"temperature"`
	SpeedX      float64 `json:"speedX"`
	SpeedY      float64 `json:"speedY"`
}

type DashboardState struct {
	mu           sync.Mutex
	ThermalImage string       `json:"thermalImage"`
	Alerts       []AlertEvent `json:"alerts"`
	CycleCount   int          `json:"cycleCount"`
	Status       string       `json:"status"`
}

var State = &DashboardState{
	Status: "Iniciando...",
	Alerts: []AlertEvent{},
}

func (s *DashboardState) UpdateImage(imagePath string) {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return
	}
	s.mu.Lock()
	s.ThermalImage = base64.StdEncoding.EncodeToString(data)
	s.mu.Unlock()
}

func (s *DashboardState) AddAlert(alert AlertEvent) {
	s.mu.Lock()
	s.Alerts = append([]AlertEvent{alert}, s.Alerts...)
	if len(s.Alerts) > 20 {
		s.Alerts = s.Alerts[:20]
	}
	s.mu.Unlock()
}

func (s *DashboardState) SetStatus(status string) {
	s.mu.Lock()
	s.Status = status
	s.mu.Unlock()
}

func Start() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/state", handleState)
	fmt.Println("Dashboard disponible en http://localhost:8080")
	go http.ListenAndServe(":8080", nil)
}

func handleState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	State.mu.Lock()
	json.NewEncoder(w).Encode(State)
	State.mu.Unlock()
}

func Timestamp() string {
	return time.Now().Format("15:04:05")
}

var htmlPage = `<!DOCTYPE html>
<html lang="es">
<head>
<meta charset="UTF-8">
<title>Perimeter Watch - FLIR PT-Series HD</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{background:#0a0e1a;color:#e0e6f0;font-family:'Courier New',monospace}
.header{background:linear-gradient(135deg,#0d1b2a,#1a2940);border-bottom:1px solid #1e3a5f;padding:16px 24px;display:flex;align-items:center}
.header h1{font-size:20px;color:#00d4ff;letter-spacing:2px}
.subtitle{font-size:11px;color:#4a7fa5;letter-spacing:1px;margin-top:4px}
.dot{width:10px;height:10px;border-radius:50%;background:#00ff88;box-shadow:0 0 8px #00ff88;animation:pulse 1.5s infinite;margin-left:auto}
@keyframes pulse{0%,100%{opacity:1}50%{opacity:0.4}}
.grid{display:grid;grid-template-columns:1fr 380px;gap:16px;padding:16px;height:calc(100vh - 72px)}
.panel{background:#0d1b2a;border:1px solid #1e3a5f;border-radius:8px;display:flex;flex-direction:column;overflow:hidden}
.panel-title{padding:10px 16px;font-size:11px;color:#4a7fa5;letter-spacing:2px;border-bottom:1px solid #1e3a5f;display:flex;justify-content:space-between;align-items:center}
.camera-feed{flex:1;display:flex;align-items:center;justify-content:center;padding:16px;position:relative;background:#060d18}
.camera-feed img{max-width:100%;max-height:100%;border-radius:4px;border:1px solid #1e3a5f}
.no-feed{color:#2a4a6a;font-size:13px;text-align:center}
.crosshair{position:absolute;top:50%;left:50%;transform:translate(-50%,-50%);width:50px;height:50px;pointer-events:none}
.crosshair::before,.crosshair::after{content:'';position:absolute;background:rgba(0,212,255,0.3)}
.crosshair::before{width:1px;height:100%;left:50%}
.crosshair::after{height:1px;width:100%;top:50%}
.right{display:flex;flex-direction:column;gap:16px}
.stats{display:grid;grid-template-columns:1fr 1fr;gap:8px;padding:12px}
.stat{background:#0a1628;border:1px solid #1e3a5f;border-radius:6px;padding:10px 12px}
.stat-label{font-size:9px;color:#4a7fa5;letter-spacing:1px;margin-bottom:4px}
.stat-value{font-size:22px;color:#00d4ff;font-weight:bold}
.stat-value.red{color:#ff4444}
.stat-value.green{color:#00ff88}
.alerts-panel{flex:1;display:flex;flex-direction:column;overflow:hidden}
.alerts-list{flex:1;overflow-y:auto;padding:8px}
.alert-item{background:#0a1628;border:1px solid #1e3a5f;border-left:3px solid #ff4444;border-radius:4px;padding:8px 12px;margin-bottom:6px;font-size:11px}
.alert-time{color:#4a7fa5;margin-bottom:3px}
.alert-data{color:#e0e6f0;margin-bottom:2px}
.alert-temp{color:#ff8800;font-weight:bold}
.alert-ptz{color:#4a7fa5}
.progress-panel{padding:12px 16px}
.prog-label{font-size:10px;color:#4a7fa5;letter-spacing:1px;margin-bottom:8px}
.prog-bar-bg{background:#0a1628;border-radius:4px;height:6px;overflow:hidden}
.prog-bar{height:100%;background:linear-gradient(90deg,#00d4ff,#00ff88);border-radius:4px;width:100%}
.status-txt{font-size:11px;color:#4a7fa5;margin-top:8px}
.badge{display:inline-block;padding:2px 8px;border-radius:10px;font-size:9px;letter-spacing:1px}
.badge.live{background:#ff444422;color:#ff4444;border:1px solid #ff444444}
.badge.ok{background:#00ff8822;color:#00ff88;border:1px solid #00ff8844}
</style>
</head>
<body>
<div class="header">
  <div>
    <h1>PERIMETER WATCH</h1>
    <div class="subtitle">FLIR PT-SERIES HD &nbsp; THERMAL DETECTION &nbsp; ONVIF 2.0 &nbsp; GO</div>
  </div>
  <div class="dot"></div>
</div>
<div class="grid">
  <div class="panel">
    <div class="panel-title">
      <span>THERMAL FEED - IR SENSOR 640x480</span>
      <span id="feedTime">--:--:--</span>
    </div>
    <div class="camera-feed">
      <img id="thermalImg" src="" style="display:none"/>
      <div class="no-feed" id="noFeed">Esperando stream termico...</div>
      <div class="crosshair"></div>
    </div>
  </div>
  <div class="right">
    <div class="panel">
      <div class="panel-title"><span>ESTADO DEL SISTEMA</span></div>
      <div class="stats">
        <div class="stat">
          <div class="stat-label">CICLO</div>
          <div class="stat-value" id="cycleVal">0</div>
        </div>
        <div class="stat">
          <div class="stat-label">ALERTAS</div>
          <div class="stat-value red" id="alertCount">0</div>
        </div>
        <div class="stat">
          <div class="stat-label">PTZ</div>
          <div class="stat-value green" id="ptzStatus">HOME</div>
        </div>
        <div class="stat">
          <div class="stat-label">SENSOR IR</div>
          <div class="stat-value green">ON</div>
        </div>
      </div>
    </div>
    <div class="panel">
      <div class="progress-panel">
        <div class="prog-label">ESTADO DEL SISTEMA</div>
        <div class="prog-bar-bg">
          <div class="prog-bar" id="progBar"></div>
        </div>
        <div class="status-txt" id="statusTxt">Iniciando sistema...</div>
      </div>
    </div>
    <div class="panel alerts-panel">
      <div class="panel-title">
        <span>REGISTRO DE ALERTAS</span>
        <span class="badge ok" id="modeBadge">PATROL</span>
      </div>
      <div class="alerts-list" id="alertsList">
        <div style="color:#2a4a6a;font-size:12px;padding:16px;text-align:center">
          Sin alertas registradas
        </div>
      </div>
    </div>
  </div>
</div>
<script>
function buildAlert(a) {
  var div = document.createElement('div');
  div.className = 'alert-item';
  div.innerHTML =
    '<div class="alert-time">' + a.timestamp + '</div>' +
    '<div class="alert-data">Centro: (' + a.centerX + ', ' + a.centerY + ') ' + a.width + 'x' + a.height + 'px</div>' +
    '<div class="alert-temp">Temp estimada: ' + a.temperature.toFixed(1) + 'C</div>' +
    '<div class="alert-ptz">PTZ pan:' + a.speedX.toFixed(2) + ' tilt:' + a.speedY.toFixed(2) + '</div>';
  return div;
}

async function update() {
  try {
    var res = await fetch('/state');
    var data = await res.json();

    if (data.thermalImage) {
      var img = document.getElementById('thermalImg');
      img.src = 'data:image/jpeg;base64,' + data.thermalImage;
      img.style.display = 'block';
      document.getElementById('noFeed').style.display = 'none';
      document.getElementById('feedTime').textContent = new Date().toLocaleTimeString();
    }

    document.getElementById('cycleVal').textContent = data.cycleCount || 0;
    document.getElementById('statusTxt').textContent = data.status || '';

    var alerts = data.alerts || [];
    document.getElementById('alertCount').textContent = alerts.length;

    if (alerts.length > 0) {
      document.getElementById('ptzStatus').textContent = 'TRACK';
      document.getElementById('ptzStatus').className = 'stat-value red';
      document.getElementById('modeBadge').textContent = 'TRACKING';
      document.getElementById('modeBadge').className = 'badge live';
      var list = document.getElementById('alertsList');
      list.innerHTML = '';
      for (var i = 0; i < alerts.length; i++) {
        list.appendChild(buildAlert(alerts[i]));
      }
    } else {
      document.getElementById('ptzStatus').textContent = 'PATROL';
      document.getElementById('ptzStatus').className = 'stat-value green';
      document.getElementById('modeBadge').textContent = 'PATROL';
      document.getElementById('modeBadge').className = 'badge ok';
    }
  } catch(e) {
    document.getElementById('statusTxt').textContent = 'Reconectando...';
  }
}

setInterval(update, 1000);
update();
</script>
</body>
</html>`

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, htmlPage)
}
