import ReactDom from "react-dom/client";
import "./index.css";
import App from "./App.jsx";
import { BrowserRouter } from "react-router-dom";

const root = document.getElementById("root");

ReactDom.createRoot(root).render(
  <BrowserRouter>
    <App />
  </BrowserRouter>
);
