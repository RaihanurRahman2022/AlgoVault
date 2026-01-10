# AlgoVault - Practice Problem Tracker

A full-stack web application for organizing and tracking coding practice problems, grouped by categories and patterns.

## Features

- 🔐 **JWT Authentication** - Secure login and registration
- 📁 **Hierarchical Organization** - Category → Pattern → Problem structure
- 💾 **Persistent Storage** - All data saved in backend database
- 📝 **Rich Problem Details** - Description, input/output, constraints, samples
- 💻 **Multi-language Solutions** - Support for C++, Go, Python, Java, JavaScript
- 📚 **Notes Section** - Keep your insights and tips
- ✏️ **Full CRUD Operations** - Create, read, update, delete all entities
- 🎨 **Modern UI** - Beautiful dark theme with Tailwind CSS

## Tech Stack

### Frontend
- React 19 + TypeScript
- Vite
- Tailwind CSS
- Lucide React (icons)

### Backend
- Go (Golang)
- SQLite (development) / PostgreSQL (production)
- JWT authentication
- Gorilla Mux router

## Project Structure

```
algovault---practice-problem-tracker/
├── frontend/              # React frontend
│   ├── components/       # Reusable components
│   ├── pages/            # Page components
│   ├── services/         # API services
│   └── types.ts          # TypeScript types
│
├── backend/              # Go backend
│   ├── main.go          # Server entry point
│   ├── handlers.go      # API handlers
│   ├── models.go        # Database models
│   └── middleware.go    # Auth middleware
│
└── DEPLOYMENT.md        # Deployment guide
```

## Getting Started

### Prerequisites

- Node.js 18+ and npm
- Go 1.21+
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone <your-repo-url>
   cd algovault---practice-problem-tracker
   ```

2. **Setup Backend**
   ```bash
   cd backend
   go mod download
   ```

3. **Setup Frontend**
   ```bash
   cd frontend
   npm install
   ```

### Running Locally

1. **Start Backend** (from `backend/` directory)
   ```bash
   go run .
   ```
   Backend runs on `http://localhost:8080`

2. **Start Frontend** (from `frontend/` directory)
   ```bash
   npm run dev
   ```
   Frontend runs on `http://localhost:3000`

3. **Access the Application**
   - Open `http://localhost:3000` in your browser
   - Register a new account or login

## Usage

### Creating Your First Problem

1. **Create a Category**
   - Click "New Category" on the dashboard
   - Enter name, icon, and description
   - Example: "Arrays & Strings", icon: "Code2"

2. **Create a Pattern**
   - Select a category
   - Click "Add Pattern"
   - Enter pattern details
   - Example: "Two Pointers", icon: "Target"

3. **Create a Problem**
   - Select a pattern
   - Click "New Problem"
   - Fill in all problem details:
     - Title, difficulty
     - Description (markdown)
     - Input/Output format
     - Constraints
     - Sample input/output with explanation
     - Solutions in multiple languages
     - Notes

### Data Structure

- **Category**: Groups related patterns
  - Name, icon, description
  - Shows count of patterns

- **Pattern**: Groups related problems
  - Name, icon, description
  - Belongs to a category
  - Shows count of problems

- **Problem**: Individual coding problem
  - Title, difficulty
  - Full problem description (markdown)
  - Input/output specifications
  - Constraints
  - Sample cases with explanations
  - Multiple language solutions
  - Personal notes

## API Endpoints

### Authentication
- `POST /api/login` - Login
- `POST /api/register` - Register

### Categories
- `GET /api/categories` - List all categories
- `POST /api/categories` - Create category
- `PUT /api/categories/{id}` - Update category
- `DELETE /api/categories/{id}` - Delete category

### Patterns
- `GET /api/categories/{categoryId}/patterns` - List patterns
- `POST /api/categories/{categoryId}/patterns` - Create pattern
- `PUT /api/patterns/{id}` - Update pattern
- `DELETE /api/patterns/{id}` - Delete pattern

### Problems
- `GET /api/patterns/{patternId}/problems` - List problems
- `GET /api/problems/{id}` - Get problem details
- `POST /api/patterns/{patternId}/problems` - Create problem
- `PUT /api/problems/{id}` - Update problem
- `DELETE /api/problems/{id}` - Delete problem

All endpoints except login/register require JWT authentication.

## Environment Variables

### Backend
- `JWT_SECRET` - Secret key for JWT tokens (required)
- `DATABASE_URL` - Database connection string (optional, defaults to SQLite)
- `PORT` - Server port (default: 8080)

### Frontend
- `VITE_API_BASE_URL` - Backend API URL (default: http://localhost:8080/api)

## Deployment

See [DEPLOYMENT.md](./DEPLOYMENT.md) for detailed deployment instructions.

**Quick Deploy Options:**
- Frontend: Vercel, Netlify, GitHub Pages
- Backend: Railway, Render, Fly.io
- Database: Supabase, Neon, Railway PostgreSQL

## Development

### Backend Development
```bash
cd backend
go run . -db ./algovault.db -port 8080 -jwt-secret your-secret
```

### Frontend Development
```bash
cd frontend
npm run dev
```

### Building for Production

**Backend:**
```bash
cd backend
go build -o server
./server
```

**Frontend:**
```bash
cd frontend
npm run build
# Output in frontend/dist
```

## Features Implemented

✅ User authentication (login/register)  
✅ JWT token-based authentication  
✅ Protected API routes  
✅ Category management (CRUD)  
✅ Pattern management (CRUD)  
✅ Problem management (CRUD)  
✅ Multi-language solution support  
✅ Markdown support for problem descriptions  
✅ Notes section for each problem  
✅ Responsive UI  
✅ Database persistence  

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

MIT License - feel free to use this project for your own practice!

## Support

For issues and questions, please open an issue on GitHub.

---

**Built with ❤️ for coding practice and interview preparation**
