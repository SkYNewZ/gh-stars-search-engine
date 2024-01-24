export type Result = {
  cost: number;
  facets: null;
  hits: Hit[];
  max_score: number;
  request: Request;
  status: Status;
  took: number;
  total_hits: number;
};

type Hit = {
  fields: Fields;
  id: string;
  index: string;
  score: number;
  sort: string[];
};

type Fields = {
  description: string;
  name_with_owner: string;
  "primary_language.color": string;
  "primary_language.name": string;
  url: string;
};

type Request = {
  explain: boolean;
  facets: null;
  fields: string[];
  from: number;
  highlight: null;
  includeLocations: boolean;
  query: Query;
  search_after: null;
  search_before: null;
  size: number;
  sort: string[];
};

type Query = {
  query: string;
};

type Status = {
  failed: number;
  successful: number;
  total: number;
};
