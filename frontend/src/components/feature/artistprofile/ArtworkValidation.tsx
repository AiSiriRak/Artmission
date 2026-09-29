export interface ArtworkFormData {
  name: string;
  category: string;
  styles: string[];
  description: string;
  minimum_deadline_days: number | string;
  price: number | string;
}

export interface ArtworkFormErrors {
  name: string;
  description: string;
  category: string;
  styles: string;
  minimum_deadline_days: string;
  price: string;
}

export const ValidateArtworkForm = (formData: ArtworkFormData): ArtworkFormErrors => {
  const errors: ArtworkFormErrors = {
    name: "",
    description: "",
    category: "",
    styles: "",
    minimum_deadline_days: "",
    price: "",
  };

  if (!formData.name.trim()) {
    errors.name = "Please enter artwork name";
  }

  if (!formData.description.trim()) {
    errors.description = "Please enter description";
  }

  if (!formData.category || formData.category.trim() === "") {
    errors.category = "Please select a category";
  }

  if (!formData.styles || formData.styles.length === 0) {
    errors.styles = "Please select at least one style";
  }

  // Deadline
  if (String(formData.minimum_deadline_days).trim() === "") {
    errors.minimum_deadline_days = "Please enter deadline";
  } else if (Number(formData.minimum_deadline_days) <= 0) {
    errors.minimum_deadline_days = "Deadline must be greater than 0";
  }

  // Price
  if (String(formData.price).trim() === "") {
    errors.price = "Please enter price";
  } else if (Number(formData.price) < 0) {
    errors.price = "Price cannot be negative";
  }

  return errors;
};;