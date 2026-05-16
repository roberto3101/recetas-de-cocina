import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";

import AplicacionEnrutada from "@/aplicacion/enrutamiento/AplicacionEnrutada";
import "@/aplicacion/estilos/index.css";

ReactDOM.createRoot(document.getElementById("raiz") as HTMLElement).render(
  <React.StrictMode>
    <BrowserRouter>
      <AplicacionEnrutada />
    </BrowserRouter>
  </React.StrictMode>
);
