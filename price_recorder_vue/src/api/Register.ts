import instance from "@/utils/request.ts";

export async function sendEmailCode(email: string): Promise<{ flag: boolean; message: string }> {
  return await instance.post("/user/send-email-code/v1", { email }) as { flag: boolean; message: string };
}

export interface RegisterReq {
  email: string
  code: string
  username: string
  password: string
}

export interface RegisterResponse {
  token: string
  userId?: number | string
  username?: string
  [key: string]: any
}

export async function registerWithEmail(data: RegisterReq): Promise<RegisterResponse> {
  return await instance.post("/user/register/v1", data) as RegisterResponse;
}
