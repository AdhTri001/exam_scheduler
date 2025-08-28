# Exam Scheduler

A browser-based exam scheduling tool that runs entirely in your browser - no server required. This application helps academic institutions schedule exams by automatically assigning courses to time slots and halls while avoiding student conflicts.

🌐 **[Live Demo](https://adhtri001.github.io/exam_scheduler/)**

## Features

- **🏠 Fully Client-Side**: All processing happens in your browser - your data never leaves your device
- **📊 CSV Import/Export**: Import student registrations and hall data from CSV files
- **🧠 Smart Scheduling**: Uses graph coloring algorithms to avoid student conflicts
- **🏛️ Hall Management**: Automatically assigns halls based on capacity and availability
- **⚡ WebAssembly Powered**: High-performance Go backend compiled to WebAssembly
- **📱 Modern UI**: Clean, responsive interface built with React and Material-UI
- **💾 Session Persistence**: Your work is automatically saved to browser storage

## Quick Start

1. **Upload Data**: Import your `registrations.csv` (student enrollments) and `halls.csv` (available venues)
2. **Configure Parameters**: Set exam dates, time slots, and scheduling preferences
3. **Generate Schedule**: Let the algorithm create an optimal exam schedule
4. **Review & Download**: Validate results and export your schedule

## Data Format

### Registrations CSV
```csv
student_id,course_id
john_doe,CS101
jane_smith,CS101
john_doe,MATH201
```

### Halls CSV
```csv
hall,capacity,group
Room_A,100,Science
Room_B,50,Arts
Auditorium,300,Large
```

### Output Schedule CSV
```csv
course_id,slot_id,slot_datetime,halls,enrolled_count,notes
CS101,2025-01-20T09:00Z#1,2025-01-20T09:00:00Z,Room_A;Room_B,85,
MATH201,2025-01-20T13:00Z#2,2025-01-20T13:00:00Z,Auditorium,120,
```

## Architecture

This project consists of two main components:

### Go Backend (`go/`)
- **WebAssembly Module**: Compiled Go code that runs in the browser
- **DSATUR Algorithm**: Graph coloring algorithm for conflict-free scheduling
- **CSV Processing**: Robust parsing with support for quoted fields and edge cases
- **Constraint Solving**: Handles hall capacity, student conflicts, and time preferences

### React Frontend (`web/`)
- **Vite + TypeScript**: Modern build setup with hot reloading
- **Material-UI**: Professional, accessible component library
- **Web Workers**: Non-blocking computation for large datasets
- **Local Storage**: Automatic session persistence

## Development

### Prerequisites
- Go 1.22+
- Node.js 18+
- Modern browser with WebAssembly support

### Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/adhtri001/exam_scheduler.git
   cd exam_scheduler
   ```

2. **Build the WebAssembly module**:
   ```bash
   cd go
   GOOS=js GOARCH=wasm go build -o ../web/public/main.wasm ./cmd/wasm
   ```

3. **Install and run the web application**:
   ```bash
   cd ../web
   npm install
   npm run dev
   ```

4. **Open your browser** to `http://localhost:5173`

### Building for Production

1. **Build WASM**:
   ```bash
   cd go
   GOOS=js GOARCH=wasm go build -o ../web/public/main.wasm ./cmd/wasm
   ```

2. **Build web app**:
   ```bash
   cd ../web
   npm run build
   ```

3. **Deploy**: The `web/dist/` directory contains the static site ready for deployment.

### Testing

- **Go tests**: `cd go && go test ./...`
- **Web linting**: `cd web && npm run lint`

## Configuration Options

- **Exam Period**: Start and end dates for the examination period
- **Slots Per Day**: Number of exam slots per day (default: 2)
- **Slot Duration**: Length of each exam in minutes (default: 180)
- **Holidays**: Dates to exclude from scheduling
- **Minimum Gap**: Minimum time between exams for the same student
- **Attempts**: Number of optimization attempts (higher = better results, slower)

## Algorithm Details

The scheduler uses a **DSATUR (Degree of Saturation)** graph coloring algorithm:

1. **Conflict Graph**: Creates a graph where courses are nodes and edges represent student conflicts
2. **Coloring**: Assigns time slots (colors) to courses while avoiding conflicts
3. **Hall Assignment**: Packs courses into available halls based on enrollment and capacity
4. **Optimization**: Runs multiple attempts with different random seeds to find the best solution

## Privacy & Security

- **No Data Upload**: All processing happens locally in your browser
- **No Tracking**: No analytics or user data collection
- **Open Source**: Full source code available for audit

## Browser Compatibility

- Chrome 69+
- Firefox 63+
- Safari 14+
- Edge 79+

(WebAssembly and modern JavaScript features required)

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes and add tests
4. Commit: `git commit -am 'Add feature'`
5. Push: `git push origin feature-name`
6. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- Built with [Go](https://golang.org/) and [React](https://reactjs.org/)
- Uses [DSATUR algorithm](https://en.wikipedia.org/wiki/DSatur) for graph coloring
- CSV parsing powered by [GoCSV](https://github.com/gocarina/gocsv) and [PapaParse](https://www.papaparse.com/)
