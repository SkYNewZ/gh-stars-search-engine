import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { useEffect, useState } from "react";
import Loader from "../components/Loader";
import githubLogo from "../assets/github.svg";
import { Result } from "../types";

const SIZE = 10;

const Results = () => {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [result, setResult] = useState<Result>();
  const query = searchParams.get("q");
  const page = searchParams.get("p");
  const from = (Number(page) - 1) * SIZE;

  useEffect(() => {
    if (query && page) {
      fetch(`/search?q=${query}&from=${from}&size=${SIZE}`)
        .then((response) => response.json())
        .then((result) => setResult(result));
    } else {
      navigate("/");
    }
  }, [navigate, from, query, page]);

  const handlePagination = (type: "previous" | "next"): void => {
    const newPage = type === "previous" ? Number(page) - 1 : Number(page) + 1;
    setSearchParams({ q: query!, p: String(newPage) });
  };

  const roundTook = (took: number): string => {
    if (took < 1000 * 1000) {
      return "less than 1ms";
    }

    if (took < 1000 * 1000 * 1000) {
      return "" + Math.round(took / (1000*1000)) + "ms";
    }

    const roundMs = Math.round(took / (1000 * 1000));
    return "" + roundMs / 1000 + "s";
  };

  return (
    <div className="container px-4 mt-8">
      <Link className="flex items-center space-x-2 mb-4 w-fit" to="/">
        <img src="/vite.svg" alt="Logo" />
        <h1 className="text-xl">Search Engine</h1>
      </Link>
      {result ? (
        <div className="overflow-hidden rounded-md border border-gray-300 bg-white my-6">
          <ul role="list" className="divide-y divide-gray-300">
            {result.hits.map((hit) => (
              <li key={hit.id} className="px-6 py-4 space-y-2">
                {/* Title */}
                <a
                  className="flex items-center space-x-2 w-fit"
                  href={hit.fields.url}
                  target="_blank"
                >
                  <img
                    src={githubLogo}
                    width="20"
                    alt={hit.fields.name_with_owner}
                  />
                  <h2 className="font-bold truncate">
                    {hit.fields.name_with_owner}
                  </h2>
                  <span
                    className="inline-flex items-center gap-x-1.5 rounded-md px-1.5 py-0.5 text-xs font-medium text-white"
                    style={{
                      backgroundColor: hit.fields["primary_language.color"],
                    }}
                  >
                    <svg
                      className="h-1.5 w-1.5 fill-white"
                      viewBox="0 0 6 6"
                      aria-hidden="true"
                    >
                      <circle cx={3} cy={3} r={3} />
                    </svg>
                    {hit.fields["primary_language.name"]}
                  </span>
                </a>
                <p>{hit.fields.description}</p>
              </li>
            ))}
          </ul>
          <nav
            className="flex items-center justify-between border-t border-gray-200 bg-white px-4 py-3 sm:px-6"
            aria-label="Pagination"
          >
            <div className="hidden sm:block">
              <p className="text-sm text-gray-700">
                Showing <span className="font-medium">{from}</span> to{" "}
                <span className="font-medium">{from + SIZE}</span> of{" "}
                <span className="font-medium">{result.total_hits}</span> results
                (
                <span className="font-medium">
                  {roundTook(result.took)}
                </span>{" "}
                seconds)
              </p>
            </div>
            <div className="flex flex-1 justify-between sm:justify-end space-x-3">
              <button
                type="button"
                className="relative inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus-visible:outline-offset-0 disabled:bg-gray-100 disabled:cursor-not-allowed"
                onClick={() => handlePagination("previous")}
                disabled={page === "1"}
              >
                Previous
              </button>
              <button
                type="button"
                className="relative inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus-visible:outline-offset-0 disabled:cursor-not-allowed"
                onClick={() => handlePagination("next")}
              >
                Next
              </button>
            </div>
          </nav>
        </div>
      ) : (
        <Loader className="flex justify-center my-8" />
      )}
    </div>
  );
};

export default Results;
