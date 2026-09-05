// === Temp Fake API Calling ===

import users from "@/data/users.json";
import { User } from "@/types/user";

// Get User Data
export async function getUser(id: number): Promise<User> {
  const user = users.find((user) => user.id === id);

  if (!user) {
    throw new Error("User not found");
  }
  console.log("Get User Data");
  return user;
}

// Update User Data
export async function updateUser(id: number, data: User): Promise<User> {
  console.log("Update User ID:", id);
  console.log("User Data Update to:", data);

  return data;
}

// Delete User Data
export async function deleteUser(id: number, data: User): Promise<User> {
  console.log("Delete User ID:", id);

  return data;
}
