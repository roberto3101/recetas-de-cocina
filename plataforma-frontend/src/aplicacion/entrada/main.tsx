import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";

import AplicacionEnrutada from "@/aplicacion/enrutamiento/AplicacionEnrutada";
import { ProveedorSesion } from "@/plataforma/identidad/usar_sesion";
import "@/aplicacion/estilos/index.css";

ReactDOM.createRoot(document.getElementById("raiz") as HTMLElement).render(
  <React.StrictMode>
    <BrowserRouter>
      <ProveedorSesion>
        <AplicacionEnrutada />
      </ProveedorSesion>
    </BrowserRouter>
  </React.StrictMode>
);
