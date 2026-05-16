import { Link } from "react-router-dom";

export default function PaginaSinResultados() {
  return (
    <main className="mx-auto max-w-2xl px-4 py-16 text-center">
      <h1 className="text-3xl font-bold text-cocina-marron">Sin resultados</h1>
      <p className="mt-3 text-cocina-oscuro">
        No encontramos recetas con ese ingrediente. Prueba con otro término o vuelve al{" "}
        <Link to="/" className="text-cocina-marron underline">
          inicio
        </Link>
        .
      </p>
    </main>
  );
}
