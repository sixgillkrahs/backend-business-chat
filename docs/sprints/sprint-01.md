# Sprint 01: Hệ Thống Authentication & Authorization (RBAC)

## 📅 Thông tin chung
* **Thời gian**: [DD/MM/YYYY] - [DD/MM/YYYY] (2 tuần)
* **Trạng thái**: 🟡 Đang diễn ra
* **Sprint Goal**: Xây dựng hoàn chỉnh luồng đăng ký, đăng nhập, quên/reset mật khẩu và cơ chế phân quyền (RBAC) từ Backend đến Frontend.

---

## 🎯 Mục tiêu chính (Sprint Goals)
- [ ] Hoàn thành toàn bộ API Auth (Login, Register, Forgot/Reset Password) ở Backend.
- [ ] Thiết kế cơ sở dữ liệu và viết Middleware phân quyền dựa trên Role (RBAC) ở Backend.
- [ ] Thiết kế giao diện (UI) và tích hợp luồng Auth ở Frontend (Web App).
- [ ] Bảo vệ các tuyến đường (Protected Routes) ở Frontend dựa trên Role của User.

---

## 📊 Bảng công việc (Task Board)

### 🖥️ 1. Backend (API & Database)

| ID | Nhiệm Vụ | Người Nhận | Độ Ưu Tiên | Trạng Thái | Pull Request / Ghi Chú |
| :--- | :--- | :--- | :---: | :---: | :--- |
| **#BE-01** | Thiết kế DB Schema cho User, Role, Permission | @dev-be | 🔥 Cao | ⚪ Todo | Bảng `users`, `roles`, `permissions` |
| **#BE-02** | Viết API Đăng ký tài khoản (Register) | @dev-be | 🔥 Cao | ⚪ Todo | Mã hóa mật khẩu (bcrypt) |
| **#BE-03** | Viết API Đăng nhập (Login) phát hành JWT | @dev-be | 🔥 Cao | ⚪ Todo | Trả về Access Token & Refresh Token |
| **#BE-04** | Xây dựng Middleware phân quyền (RBAC Middleware) | @dev-be | 🔥 Cao | ⚪ Todo | Kiểm tra Role/Permission của User |
| **#BE-05** | Viết API Yêu cầu quên mật khẩu (Forgot Password) | @dev-be | ⚡ Trung bình | ⚪ Todo | Tạo mã OTP/Token và gửi qua Email |
| **#BE-06** | Viết API Đặt lại mật khẩu (Reset Password) | @dev-be | ⚡ Trung bình | ⚪ Todo | Xác thực OTP/Token & cập nhật DB |

### 🎨 2. Frontend (Giao diện & Tích hợp)

| ID | Nhiệm Vụ | Người Nhận | Độ Ưu Tiên | Trạng Thái | Pull Request / Ghi Chú |
| :--- | :--- | :--- | :---: | :---: | :--- |
| **#FE-01** | Thiết kế Màn hình Đăng nhập (Login Page) | @dev-fe | 🔥 Cao | ⚪ Todo | Validate form, xử lý lỗi |
| **#FE-02** | Thiết kế Màn hình Đăng ký (Register Page) | @dev-fe | 🔥 Cao | ⚪ Todo | Validate dữ liệu nhập vào |
| **#FE-03** | Thiết kế Màn hình Quên mật khẩu | @dev-fe | ⚡ Trung bình | ⚪ Todo | Nhập email nhận mã |
| **#FE-04** | Thiết kế Màn hình Đặt lại mật khẩu (Reset Password) | @dev-fe | ⚡ Trung bình | ⚪ Todo | Nhập OTP và mật khẩu mới |
| **#FE-05** | Tích hợp Auth State Management & API Client | @dev-fe | 🔥 Cao | ⚪ Todo | Lưu JWT vào Cookie/LocalStorage |
| **#FE-06** | Xây dựng Protected Routes phân quyền theo Role | @dev-fe | 🔥 Cao | ⚪ Todo | Redirect user nếu không đủ quyền truy cập |

---

## 🛠️ Rủi ro & Khó khăn (Risks & Blockers)
* **Tích hợp Email**: Cần chuẩn bị sẵn tài khoản SMTP (như Gmail SMTP hoặc SendGrid) để Backend gửi mã OTP/Reset link.
* **Đồng bộ cơ sở dữ liệu**: Đảm bảo migration DB được thực hiện chính xác ở cả local và staging.

---

## 📝 Tổng kết Sprint (Retro & Review)
*(Sẽ được điền vào ngày cuối cùng của Sprint)*
* **Đạt được**: ...
* **Cần cải thiện**: ...
* **Hành động (Action Items)**:
  - [ ] Action 1...
