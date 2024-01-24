import { MagnifyingGlassIcon } from "@heroicons/react/20/solid";
import viteLogo from "/vite.svg";
import { useNavigate } from "react-router-dom";

const Home = () => {
  const navigate = useNavigate();

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const query = form.get("query");
    navigate(`/search?q=${query}&p=1`);
  };

  return (
    <form className="mt-36 md:mt-64 flex flex-col" onSubmit={handleSubmit}>
      <img className="mx-auto mb-4" src={viteLogo} alt="Logo" width="120" />
      <h1 className="mb-8 text-xl text-center">Search Engine</h1>
      <div className="mx-auto w-4/5 md:w-1/2 max-w-lg mb-4 relative">
        <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
          <MagnifyingGlassIcon
            className="h-5 w-5 text-gray-400"
            aria-hidden="true"
          />
        </div>
        <input
          type="text"
          className="block w-full rounded-md border-0 py-1.5 pl-10 text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
          name="query"
        />
      </div>
      <button
        type="submit"
        className="mx-auto w-48 rounded-md bg-white px-2.5 py-1.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
      >
        Search
      </button>
    </form>
  );
};

export default Home;
