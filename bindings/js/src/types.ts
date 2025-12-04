// Mirrors the Go-side minifyOptions struct in minify.go.
type LiteralUnion<T extends string, U extends string> = T | (U & Record<never, never>);

// mirrors regex from the Go library, allows for blabla/json, blabla+json, blabla/xml, etc.
type CustomMinifyMediaType =
  | `${string}script${string}`
  | `${string}/json`
  | `${string}+json`
  | `${string}/xml`
  | `${string}+xml`;


// mirrors the known media types from the Go library
type KnownMinifyMediaType =
  | 'importmap'
  | 'speculationrules'
  | 'text/css'
  | 'text/html'
  | 'image/svg+xml'
  | 'text/asp'
  | 'text/x-ejs-template'
  | 'application/x-httpd-php'
  | 'text/x-template'
  | 'text/x-go-template'
  | 'text/x-mustache-template'
  | 'text/x-handlebars-template'
  | 'module';

// common media types from regex that are not directly listed in the Go library
// added for better DX
type CommonMinifyMediaType =
  | 'application/javascript'
  | 'text/javascript'
  | 'application/json'
  | 'application/rss+xml'
  | 'application/manifest+json'
  | 'application/xhtml+xml'
  | 'text/xml';

export type MinifyMediaType = LiteralUnion<KnownMinifyMediaType | CommonMinifyMediaType, CustomMinifyMediaType>;

export interface MinifyConfig {
  cssPrecision?: number;
  cssVersion?: number;
  htmlKeepComments?: boolean;
  htmlKeepConditionalComments?: boolean;
  htmlKeepDefaultAttrvals?: boolean;
  htmlKeepDocumentTags?: boolean;
  htmlKeepEndTags?: boolean;
  htmlKeepQuotes?: boolean;
  htmlKeepSpecialComments?: boolean;
  htmlKeepWhitespace?: boolean;
  jsKeepVarNames?: boolean;
  jsPrecision?: number;
  jsVersion?: number;
  jsonKeepNumbers?: boolean;
  jsonPrecision?: number;
  svgKeepComments?: boolean;
  svgPrecision?: number;
  xmlKeepWhitespace?: boolean;
}

export interface MinifyOptions extends MinifyConfig {
  data: string;
  type: MinifyMediaType;
}
