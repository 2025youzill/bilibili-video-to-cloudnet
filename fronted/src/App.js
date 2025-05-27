import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import LoginPage from './pages/LoginPage';
import UploadPage from './pages/UploadPage';

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<LoginPage />} />
        <Route path="/upload" element={<UploadPage />} />
      </Routes>
    </Router>
  );
}

export default App;