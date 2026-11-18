export interface TemplateRow {
  id: string;
  typeCode: string;
  title: string;
  body: string;
  link: string;
}

export interface TemplateFormData {
  id?: string;
  typeCode: string;
  title: string;
  body: string;
  link: string;
  sample: string;
}

export interface TypeOption {
  code: string;
  name: string;
}

export function emptyTemplateForm(): TemplateFormData {
  return {
    typeCode: "",
    title: "",
    body: "",
    link: "",
    sample: '{"OrderNo":"A1001"}'
  };
}
