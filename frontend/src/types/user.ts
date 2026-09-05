export interface User {
  id: number;
  role: string;
  name: string;
  password: string;
  email: string;
  bank: {
    name: string;
    accountHolder: string;
    accountNumber: string;
  };
}
