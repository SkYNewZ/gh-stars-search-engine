import { Routes, Route } from "react-router-dom";
import Home from "./pages/Home";
import Results from "./pages/Results";

const App = () => (
  <Routes>
    <Route index path="/" element={<Home />} />
    <Route path="/search" element={<Results />} />
  </Routes>
);

export default App;
